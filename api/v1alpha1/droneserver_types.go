/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SecretRef struct {
	// Name of the Secret
	Name string `json:"name"`
	// Key within the Secret
	Key string `json:"key"`
}

// DroneServerSpec defines the desired state of DroneServer.
type DroneServerSpec struct {
	// RunnerReplicas is the number of Drone runner replicas
	// Reference to the GitHub Client ID secret
	GithubClientIDSecret SecretRef `json:"githubClientIDSecret"`
	// Reference to the GitHub Client Secret secret
	GithubClientSecretSecret SecretRef `json:"githubClientSecretSecret"`

	SharedSecret SecretRef `json:"sharedSecret"`

	RunnerReplicas int32 `json:"runnerReplicas"`
	// RunnerCapacity is the max concurrent pipelines per runner
	RunnerCapacity int32 `json:"runnerCapacity"`
	// ServerHost is the DRONE_SERVER_HOST (e.g., drone.example.com)
	ServerHost string `json:"serverHost"`
	// GithubServer is the DRONE_GITHUB_SERVER (e.g., https://github.com)
	GithubServer string `json:"githubServer"`
}

// DroneServerStatus defines the observed state of DroneServer
type DroneServerStatus struct {
	// Phase of the server (e.g., Running, Failed)
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DroneServer is the Schema for the droneservers API.
type DroneServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DroneServerSpec   `json:"spec,omitempty"`
	Status DroneServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DroneServerList contains a list of DroneServer.
type DroneServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DroneServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DroneServer{}, &DroneServerList{})
}
