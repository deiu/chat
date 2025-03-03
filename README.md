# Chat

A real-time chat application with direct messaging support, built with Go and WebSocket.

You can see a live demo at https://deiu-chat.onrender.com but you may have to wait up to a minutefor the server to restart after a long period of inactivity.

## Features

- Real-time messaging using WebSocket
- Direct user-to-user messaging
- Case-insensitive unique usernames
- Unread message notifications
- Online users list
- Docker support
- Clean disconnection handling
- Message history per conversation

## Usage

1. Open the application in your browser (default: http://localhost:8080)
2. Enter a username to login
3. Select a user from the online users list to start chatting
4. Messages are delivered in real-time
5. Unread messages are indicated with (*)
6. Use the logout button to disconnect

## Development

### Prerequisites

- Go 1.21 or later
- Docker (optional)

### Building the application

```bash
go build -o chat main.go
```

### Running the application

```bash
./chat
```

### Using Docker

1. Clone the repository:

```bash
git clone https://github.com/deiu/chat.git
cd chat
```

2. Build and run the application:

```bash
docker build -t chat .
docker run -p 8080:8080 chat
```

3. Access the application in your browser at `http://localhost:8080`.

### Using Docker Compose

1. Clone the repository:

```bash
git clone https://github.com/deiu/chat.git
cd chat
```

2. Build and run the application:

```bash
docker-compose up --build
```

3. Access the application in your browser at `http://localhost:8080`.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details