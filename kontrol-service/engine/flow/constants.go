package flow

import (
	"fmt"
	"strings"
)

var TraceHeaderPriorities = []string{
	"x-kardinal-trace-id",   // Our custom header (checked first)
	"x-b3-traceid",          // Zipkin B3
	"x-request-id",          // General request ID, often used for tracing
	"x-cloud-trace-context", // Google Cloud Trace
	"x-amzn-trace-id",       // AWS X-Ray
	"traceparent",           // W3C Trace Context
	"uber-trace-id",         // Jaeger
	"x-datadog-trace-id",    // Datadog
}

const (
	inboundRequestTraceIDFilter = `
function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local trace_id = headers:get("x-kardinal-trace-id")
  
  if not trace_id then
    request_handle:respond(
      {[":status"] = "400"},
      "Missing required x-kardinal-trace-id header"
    )
  end
end
`

	outgoingRequestTraceIDFilterTemplate = `
%s

function get_trace_id(headers)
  for _, header_name in ipairs(trace_header_priorities) do
    local trace_id = headers:get(header_name)
    if trace_id then
      return trace_id, header_name
    end
  end

  return nil, nil
end

function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local trace_id, source_header = get_trace_id(headers)
  local hostname = headers:get(":authority")
  
  if not trace_id then
    request_handle:logWarn("No valid trace ID found in request headers")
    request_handle:respond(
      {[":status"] = "400"},
      "Missing required trace ID header"
    )
    return
  end

  if source_header ~= "x-kardinal-trace-id" then
    request_handle:headers():add("x-kardinal-trace-id", trace_id)
    request_handle:logInfo("Set x-kardinal-trace-id from " .. source_header .. ": " .. trace_id)
  end

  local destination = determine_destination(request_handle, trace_id, hostname)
  request_handle:headers():add("x-kardinal-destination", destination)
end

function determine_destination(request_handle, trace_id, hostname)
  hostname = hostname:match("^([^:]+)")
  local headers, body = request_handle:httpCall(
    "outbound|8080||trace-router.default.svc.cluster.local",
    {
      [":method"] = "GET",
      [":path"] = "/route?trace_id=" .. trace_id .. "&hostname=" .. hostname,
      [":authority"] = "trace-router.default.svc.cluster.local"
    },
    "",
    5000
  )
  
  if not headers or headers[":status"] ~= "200" then
    request_handle:logWarn("Failed to determine destination, falling back to prod")
    return hostname .. "-prod"  -- Fallback to prod
  end
  
  return body
end
`

	luaFilterType = "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
)

func generateLuaTraceHeaderPriorities() string {
	var sb strings.Builder
	sb.WriteString("local trace_header_priorities = {")
	for i, header := range TraceHeaderPriorities {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", header))
	}
	sb.WriteString("}")
	return sb.String()
}
func getOutgoingRequestTraceIDFilter() string {
	return fmt.Sprintf(outgoingRequestTraceIDFilterTemplate, generateLuaTraceHeaderPriorities())
}
