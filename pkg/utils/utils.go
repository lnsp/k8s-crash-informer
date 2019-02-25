package utils

import (
	"fmt"
	"io/ioutil"
)

const namespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// Namespace returns the namespace this pod is running in.
func Namespace() (string, error) {
	nsfile, err := ioutil.ReadFile(namespaceFilePath)
	if err != nil {
		return "", fmt.Errorf("could not read namespace: %v", err)
	}
	return string(nsfile), nil
}
