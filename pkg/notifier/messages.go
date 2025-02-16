package notifier

import "log"

func handleMessage(data map[string]interface{}) {
	msgType, ok := data["type"].(string)
	if !ok {
		log.Printf("Message missing type field")
		return
	}

	switch msgType {
	case "location":
		log.Printf("Received location update: %v", data)
	case "status":
		log.Printf("Received status update: %v", data)
	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}
