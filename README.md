# go-findmy

## Project Overview

`go-findmy` is a Go-based application designed to interact with Google's Find My Device network. It allows users to retrieve location data for their registered devices. The project is structured into several key packages:

-   `cmd/`: Contains the main application entry point.
-   `internal/`: Houses the core logic of the application.
    -   `findmy/`: Manages device interactions, including fetching device information and triggering location updates.
    -   `logger/`: Provides logging functionalities.
    -   `publisher/`: Handles the publishing of device data (e.g., to an MQTT broker).
    -   `utilities.go`: Contains shared utility functions.
-   `pkg/`: Includes various modules for specific functionalities.
    -   `decryptor/`: Responsible for decrypting location data received from the Find My Device network.
    -   `notifier/`: Manages notifications and communication with Firebase Cloud Messaging (FCM) to receive device updates.
    -   `nova/`: Implements the client for interacting with Google's Find My Device network infrastructure. This includes device listing and action execution (e.g., requesting a location update).
    -   `shared/`: Contains shared models, constants, and utilities used across different packages, such as data structures for devices, locations, and Vault client for secret management.
    -   `spot/`: Appears to be another client interacting with a part of the Find My Device network.

## Key Functionalities

-   **Device Discovery**: Lists devices registered to a Google account.
-   **Location Retrieval**: Fetches the last known location of devices.
-   **Location Decryption**: Decrypts encrypted location reports.
-   **Real-time Updates**: Listens for real-time location updates via FCM.
-   **Data Publishing**: Can publish device and location data to external systems (e.g., MQTT, Home Assistant).
-   **Semantic Location Processing**: Interprets and processes semantic location names (e.g., "Home", "Work").
-   **Secure Credential Management**: Utilizes HashiCorp Vault for managing sensitive credentials.

## Core Components

1.  **FindMy Service (`internal/findmy/`)**:
    *   Orchestrates the interaction between the Nova client, Notifier client, and Decryptor.
    *   Manages a list of devices and periodically refreshes their status.
    *   Can be configured to publish device information and location updates via the `publisher` component.
    *   Uses a scheduler (gocron) to perform periodic tasks like refreshing device locations.

2.  **Nova Client (`pkg/nova/`)**:
    *   Authenticates with and communicates with Google's Find My Device servers.
    *   Retrieves a list of associated devices (`GetDevices`).
    *   Executes actions on devices, such as pinging them to report their location (`ExecuteAction`).

3.  **Notifier Client (`pkg/notifier/`)**:
    *   Establishes a connection with FCM to receive push notifications containing location updates.
    *   Manages FCM session details, including tokens and credentials.
    *   Processes incoming messages, decodes them, and passes them for decryption.

4.  **Decryptor (`pkg/decryptor/`)**:
    *   Handles the decryption of encrypted location payloads received through the notifier.
    *   Supports different decryption methods based on the presence of a public key in the report.
    *   Converts raw decrypted data into structured location reports (latitude, longitude, altitude, accuracy).

5.  **Publisher (`internal/publisher/`)**:
    *   Provides an interface to publish device data and location reports to an MQTT broker or other messaging systems.

6.  **Vault Integration (`pkg/shared/vault/`)**:
    *   Securely retrieves necessary credentials (e.g., API keys, session tokens) from a HashiCorp Vault instance.

## Future Enhancements
-   **Web Interface**: Develop a web-based dashboard for visualizing device locations and statuses.
-   **Initial Authentication**: Currently requires manual authentication via to fetch initial tokens. Future versions may implement a more automated authentication flow.