package topology

import (
	"github.com/sirupsen/logrus"
)

type RawKialiGraph struct {
	Elements struct {
		Nodes []struct {
			Data struct {
				ID       string `json:"id"`
				NodeType string `json:"nodeType"`
				Service  string `json:"service"`
				App      string `json:"app"`
				Version  string `json:"version"`
			} `json:"data"`
		} `json:"nodes"`
		Edges []struct {
			Data struct {
				Source string `json:"source"`
				Target string `json:"target"`
			} `json:"data"`
		} `json:"edges"`
	} `json:"elements"`
}

type Node struct {
	ID             string
	ServiceName    string
	ServiceVersion string
	TalksTo        []string
}

func graphToNodesMap(graph *RawKialiGraph) map[string]*Node {
	nodesMap := make(map[string]*Node)

	// Populate nodes
	for _, n := range graph.Elements.Nodes {
		serviceName := n.Data.Service
		if serviceName == "" {
			serviceName = n.Data.App // Use app name if service name is not specified
		}
		serviceVersion := n.Data.Version
		if serviceVersion == "" {
			serviceVersion = "latest" // Default to 'latest' if no version is specified
		}

		nodesMap[n.Data.ID] = &Node{
			ID:             n.Data.ID,
			ServiceName:    serviceName,
			ServiceVersion: serviceVersion,
			TalksTo:        make([]string, 0),
		}
	}

	// Populate connections
	for _, e := range graph.Elements.Edges {
		if sourceNode, ok := nodesMap[e.Data.Source]; ok {
			if targetNode, ok := nodesMap[e.Data.Target]; ok {
				sourceNode.TalksTo = append(sourceNode.TalksTo, targetNode.ID)
			}
		}
	}

	// Print the nodes and their connections
	for _, node := range nodesMap {
		logrus.Infof("Node ID: %s", node.ID)
		logrus.Infof("  Service: %s Version: %s", node.ServiceName, node.ServiceVersion)
		logrus.Infof("  Talks To: %v", node.TalksTo)
	}

	return nodesMap
}
