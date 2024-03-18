# Socket chat
This is a basic implementation of chat server and client with text-based interface.
The communication is done mainly via TCP sockets, while separate UDP and multicast
sockets are restricted to sending an ASCII art typing U and M respectively as a message.

## Running
Type
```bash
go run ./cmd/server
```
to run the server on port 9001. The client will
connect with a randomized ephemeral port on all its sockets after you run it with
```bash
go run ./cmd/client
```

This mini-project was developed as a part of the distributed systems course at AGH UST.
