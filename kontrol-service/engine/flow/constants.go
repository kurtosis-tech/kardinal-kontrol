package flow

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
	// TODO(gm) - drop fallbacks and just exit the request like you exit in inboundRequestTraceIDFilter
	outgoingRequestTraceIDFilter = `
function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  local trace_id = headers:get("x-kardinal-trace-id")
  local hostname = headers:get(":authority")
  
  if trace_id then
	local destination = determine_destination(request_handle, trace_id, hostname)
	request_handle:headers():add("x-kardinal-destination", destination)
  end
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
  
  if not headers then
	return hostname .. "-prod"  -- Fallback to prod
  end
  
  return body
end
`

	luaFilterType = "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
)
