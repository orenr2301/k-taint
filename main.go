package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
	"path/filepath"

    "k8s.io/client-go/kubernetes"
    //"k8s.io/client-go/rest"
    "k8s.io/apimachinery/pkg/types"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// NodeInfo holds information about a node and its taints
type NodeInfo struct {
    Name   string      `json:"name"`
    Taints []v1.Taint  `json:"taints"`
}

// Global variable to store nodes and their taints
var nodes []NodeInfo

// Handler to list nodes
func listNodesHandler(w http.ResponseWriter, r *http.Request) {
    // Set up Kubernetes client
    // config, err := rest.InClusterConfig()
    // if err != nil {
    //     http.Error(w, "Failed to get Kubernetes config", http.StatusInternalServerError)
    //     return
    // }

    // clientset, err := kubernetes.NewForConfig(config)
    // if err != nil {
    //     http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
    //     return
    // }
	// Load kubeconfig from the default location
    

    userHomeDir, err := os.UserHomeDir()
	if err !=nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	if userHomeDir == "" {
		fmt.Errorf("user home directory is empty")
	}

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig %s\n", kubeConfigPath)
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err !=nil {
		 fmt.Errorf("error getting kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Errorf("error getting kubernetes clientset: %v", err)
	}


    // Get the list of nodes
    nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        http.Error(w, "Failed to list nodes", http.StatusInternalServerError)
        return
    }

    // Populate nodes slice
    nodes = make([]NodeInfo, len(nodeList.Items))
    for i, node := range nodeList.Items {
        nodes[i] = NodeInfo{
            Name:   node.Name,
            Taints: node.Spec.Taints,
        }
    }

    // Return the list of nodes as JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(nodes)
}

// Handler to apply a taint to a node
func taintNodeHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        NodeName string `json:"nodeName"`
        Key      string `json:"key"`
        Value    string `json:"value"`
        Effect   string `json:"effect"` // e.g., NoSchedule, PreferNoSchedule, NoExecute
    }

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Set up Kubernetes client
    // config, err := rest.InClusterConfig()
    // if err != nil {
    //     http.Error(w, "Failed to get Kubernetes config", http.StatusInternalServerError)
    //     return
    // }

    // clientset, err := kubernetes.NewForConfig(config)
    // if err != nil {
    //     http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
    //     return
    // }

	userHomeDir, err := os.UserHomeDir()
	if err !=nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	if userHomeDir == "" {
		fmt.Errorf("user home directory is empty")
	}

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig %s\n", kubeConfigPath)
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err !=nil {
		 fmt.Errorf("error getting kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Errorf("error getting kubernetes clientset: %v", err)
	}

    // Create the taint
    taint := v1.Taint{
        Key:    input.Key,
        Value:  input.Value,
        Effect: v1.TaintEffect(input.Effect),
    }

    // Patch the node with the taint
    patchData := []byte(fmt.Sprintf(`[{"op": "add", "path": "/spec/taints", "value": %s}]`, toJSON([]v1.Taint{taint})))
    _, err = clientset.CoreV1().Nodes().Patch(context.TODO(), input.NodeName, types.JSONPatchType, patchData, metav1.PatchOptions{})
    if err != nil {
        http.Error(w, fmt.Sprintf("Error adding taint to node %s: %v", input.NodeName, err), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Taint added to node %s successfully.\n", input.NodeName)
}

// Handler to remove a taint from a node
func removeTaintHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        NodeName string `json:"nodeName"`
        Key      string `json:"key"`
    }

    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Set up Kubernetes client
    // config, err := rest.InClusterConfig()
    // if err != nil {
    //     http.Error(w, "Failed to get Kubernetes config", http.StatusInternalServerError)
    //     return
    // }

    // clientset, err := kubernetes.NewForConfig(config)
    // if err != nil {
    //     http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
    //     return
    // }
	userHomeDir, err := os.UserHomeDir()
	if err !=nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	if userHomeDir == "" {
		fmt.Errorf("user home directory is empty")
	}

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig %s\n", kubeConfigPath)
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err !=nil {
		 fmt.Errorf("error getting kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		fmt.Errorf("error getting kubernetes clientset: %v", err)
	}
    // Patch the node to remove the taint
    patchData := []byte(fmt.Sprintf(`[{"op": "remove", "path": "/spec/taints/%s"}]`, input.Key))
    _, err = clientset.CoreV1().Nodes().Patch(context.TODO(), input.NodeName, types.JSONPatchType, patchData, metav1.PatchOptions{})
    if err != nil {
        http.Error(w, fmt.Sprintf("Error removing taint from node %s: %v", input.NodeName, err), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Taint %s removed from node %s successfully.\n", input.Key, input.NodeName)
}

// Helper function to convert taint to JSON
func toJSON(taints []v1.Taint) string {
    data, _ := json.Marshal(taints)
    return string(data)
}

// Main function to start the server
func main() {

	 // Serve the index.html file at the root endpoint
	 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")  // Update path if "index.html" is in a subdirectory
    })

    http.HandleFunc("/nodes", listNodesHandler)
    http.HandleFunc("/taint-node", taintNodeHandler)
    http.HandleFunc("/remove-taint", removeTaintHandler)

    fmt.Println("Server is running on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Println("Failed to start server:", err)
        os.Exit(1)
    }
}
