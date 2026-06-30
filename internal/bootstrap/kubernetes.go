package bootstrap

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

func GetSecretData(c *HTTPClient, kubeAPI, namespace, token, name string) (map[string]string, error) {
	secretURL := fmt.Sprintf("%s/api/v1/namespaces/%s/secrets/%s", kubeAPI, namespace, name)
	payload, err := c.JSONRequest(http.MethodGet, secretURL, BearerHeaders(token), nil)
	if err != nil {
		return nil, err
	}

	var secret struct {
		Data map[string]string `json:"data"`
	}
	if err := unmarshalInto(payload, &secret); err != nil {
		return nil, err
	}

	decoded := make(map[string]string, len(secret.Data))
	for key, value := range secret.Data {
		raw, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		decoded[key] = string(raw)
	}
	return decoded, nil
}

func UpsertSecret(c *HTTPClient, kubeAPI, namespace, token, name, key, value string, labels map[string]string) error {
	secretURL := fmt.Sprintf("%s/api/v1/namespaces/%s/secrets/%s", kubeAPI, namespace, name)
	body := map[string]any{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
			"labels":    labels,
		},
		"type": "Opaque",
		"data": map[string]string{
			key: base64.StdEncoding.EncodeToString([]byte(value)),
		},
	}

	payload, err := c.JSONRequest(http.MethodGet, secretURL, BearerHeaders(token), nil)
	if err == nil {
		var existing map[string]any
		if err := unmarshalInto(payload, &existing); err != nil {
			return err
		}
		if metadata, ok := existing["metadata"].(map[string]any); ok {
			if rv, ok := metadata["resourceVersion"].(string); ok {
				body["metadata"].(map[string]any)["resourceVersion"] = rv
			}
		}
		_, err := c.JSONRequest(http.MethodPut, secretURL, BearerHeaders(token), mustMarshal(body))
		return err
	}

	_, err = c.JSONRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/namespaces/%s/secrets", kubeAPI, namespace), BearerHeaders(token), mustMarshal(body))
	return err
}

func UpsertConfigMap(c *HTTPClient, kubeAPI, namespace, token, name string, data map[string]string, labels map[string]string) error {
	configMapURL := fmt.Sprintf("%s/api/v1/namespaces/%s/configmaps/%s", kubeAPI, namespace, name)
	body := map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
			"labels":    labels,
		},
		"data": data,
	}

	payload, err := c.JSONRequest(http.MethodGet, configMapURL, BearerHeaders(token), nil)
	if err == nil {
		var existing map[string]any
		if err := unmarshalInto(payload, &existing); err != nil {
			return err
		}
		if metadata, ok := existing["metadata"].(map[string]any); ok {
			if rv, ok := metadata["resourceVersion"].(string); ok {
				body["metadata"].(map[string]any)["resourceVersion"] = rv
			}
		}
		_, err := c.JSONRequest(http.MethodPut, configMapURL, BearerHeaders(token), mustMarshal(body))
		return err
	}

	_, err = c.JSONRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/namespaces/%s/configmaps", kubeAPI, namespace), BearerHeaders(token), mustMarshal(body))
	return err
}
