package bootstrap

import "os"

func kubeServicePort() string {
	if port := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS"); port != "" {
		return port
	}
	return os.Getenv("KUBERNETES_SERVICE_PORT")
}
