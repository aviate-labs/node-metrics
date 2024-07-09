package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aviate-labs/agent-go/registry"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
)

type NodeDetails struct {
	ID string
	registry.NodeDetails
}

type DashboardNodeDetails struct {
	ID   string `json:"node_id"`
	DCID string `json:"dc_id"`
	IPv6 string `json:"ip_address"`
}

func getNodesFromDashboard(nodeProviderID string) (map[string][]NodeDetails, error) {
	resp, err := http.Get(fmt.Sprintf("https://ic-api.internetcomputer.org/api/v3/nodes?node_provider_id=%s", nodeProviderID))
	if err != nil {
		return nil, fmt.Errorf("failed to get node list from dashboard: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var wrapper struct {
		Nodes []DashboardNodeDetails `json:"nodes"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	nodeMap := make(map[string][]NodeDetails)
	for _, n := range wrapper.Nodes {
		nodeMap[nodeProviderID] = append(nodeMap[nodeProviderID], NodeDetails{
			ID:          n.ID,
			NodeDetails: registry.NodeDetails{DCID: n.DCID, IPv6: n.IPv6},
		})
	}
	return nodeMap, nil
}

func main() {
	nodeProviderID := flag.String("node-provider-id", "", "Node provider ID to export")
	useDashboard := flag.Bool("use-dashboard", false, "Use the dashboard API to get node list")
	out := flag.String("out", "prometheus.yml", "Output file")
	flag.Parse()

	if nodeProviderID == nil || *nodeProviderID == "" {
		log.Fatalln("The <node-provider-id> flag must be provided.")
	}
	if useDashboard == nil {
		*useDashboard = false
	}

	var npMap map[string][]NodeDetails
	if *useDashboard {
		var err error
		npMap, err = getNodesFromDashboard(*nodeProviderID)
		if err != nil {
			log.Fatalln("Failed to get node list from dashboard:", err)
		}
	} else {
		rc, err := registry.New()
		if err != nil {
			log.Fatalln("Failed to create registry client:", err)
		}
		log.Println("Getting node list... This may take a few minutes.")
		nm, err := rc.GetNodeListSince(0)
		if err != nil {
			log.Fatalln("Failed to get node list:", err)
		}

		npMap = make(map[string][]NodeDetails)
		for id, node := range nm {
			if v, ok := npMap[node.NodeProviderID.String()]; ok {
				npMap[node.NodeProviderID.String()] = append(v, NodeDetails{ID: id, NodeDetails: node})
			} else {
				npMap[node.NodeProviderID.String()] = []NodeDetails{{ID: id, NodeDetails: node}}
			}
		}
	}

	nodeProvider, ok := npMap[*nodeProviderID]
	if !ok {
		log.Fatalf("Failed to find node provider: %q.\n", *nodeProviderID)
	}
	log.Printf("Found %d nodes for %q.\n", len(nodeProvider), *nodeProviderID)

	var addresses []map[string]any
	for _, n := range nodeProvider {
		addresses = append(addresses, map[string]any{
			"targets": []string{fmt.Sprintf("[%s]:42372", n.IPv6)},
			"labels": map[string]string{
				"id": n.ID,
				"np": *nodeProviderID,
				"dc": n.DCID,
			},
		})
	}

	configRaw, err := yaml.Marshal(map[string]any{
		"global": map[string]any{
			"scrape_interval":     "30s",
			"evaluation_interval": "30s",
		},
		"scrape_configs": []map[string]any{
			{
				"job_name": "hostos",
				"scheme":   "https",
				"tls_config": map[string]any{
					"insecure_skip_verify": true,
				},
				"metrics_path":   "/metrics/hostos_node_exporter",
				"static_configs": addresses,
			},
			{
				"job_name": "guestos",
				"scheme":   "https",
				"tls_config": map[string]any{
					"insecure_skip_verify": true,
				},
				"metrics_path":   "/metrics/guestos_node_exporter",
				"static_configs": addresses,
			},
		},
	})
	if err != nil {
		log.Fatalln("Failed to marshal YAML:", err)
	}
	if err := os.WriteFile(*out, configRaw, 0644); err != nil {
		log.Fatalln("Failed to write YAML:", err)
	}
	log.Println("Wrote YAML to", *out)
}
