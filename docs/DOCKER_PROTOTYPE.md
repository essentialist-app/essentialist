# Essentialist Web Prototype (Docker)

This directory contains a prototype for running Essentialist as a web application using Docker.

## Overview

This prototype uses a **Client-Server architecture**:
*   **Server**: A Go backend running in the Docker container. It manages your flashcards (reading/writing to the disk) and serves the API.
*   **Client**: The Essentialist application compiled to WebAssembly (WASM), running in your browser. It communicates with the server to display and update your decks.

## How to Run

### Prerequisites
*   Docker and Docker Compose installed on your machine.

### Quick Start

1.  Navigate to the root of the repository.
2.  Run the Docker container:
    ```bash
    docker-compose up --build
    ```
3.  Open your browser to [http://localhost:8080](http://localhost:8080).

## Managing Flashcards

The Docker container expects your flashcards (`.md` files) to be in a directory mounted to `/data`.

By default, `docker-compose.yml` mounts the `./decks` directory from your host machine:

```yaml
    volumes:
      - ./decks:/data
```

To add your own decks:
1.  Create a `decks` folder in the project root (if it doesn't exist).
2.  Copy your Markdown flashcard files into it.
3.  Refresh the web page. Your new decks will appear in the list.

## Architecture Details

*   **Filesystem Access**: The web version cannot access your local hard drive directly due to browser security sandboxing. Instead, it asks the Server component to read/write files for it.
*   **Offline Support**: Currently, this prototype requires a connection to the server to fetch cards.
