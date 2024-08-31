// Copyright 2024 syumai
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

type MessageType string

const (
	MessageTypePublish MessageType = "message_type_publish"
	MessageTypeClose   MessageType = "message_type_close"
)

type Message struct {
	Type MessageType `json:"type"`
}

type ConnectionMode string

const (
	ConnectionModeSubscribe ConnectionMode = "connection_mode_subscribe"
	ConnectionModePublish   ConnectionMode = "connection_mode_publish"
)

type ConnectionMessage struct {
	ConnectionMode ConnectionMode `json:"connectionMode"`
}

type AcceptationMessage struct {
	AcceptationResult string `json:"acceptationResult"`
}
