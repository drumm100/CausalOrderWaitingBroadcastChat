# Causal Order Waiting Broadcast Chat

:beginner: Author: Andrius Lima

This project contains a implementation of the causal order waiting broadcast algorithm presented on book [Introduction to Reliable and Secure Distributed Programming - Cachin, Guerraoui, Rodrigue](https://www.springer.com/gp/book/9783642152597)

To test this implementation a chat application was created. To run do as follows:

`go run Chat.go 1 127.0.0.1:5001 127.0.0.1:6001`

`go run Chat.go 2 127.0.0.1:6001 127.0.0.1:5001`

This will create 2 chat sessions on your terminal. To send a message type some text and hit enter. After hiting enter the application will broadcast the message to every address.
If the message contains the keyword `delay` the message will take 10 seconds to be sended by the link layer.

## Arguments

`go run Chat.go PROCCESS INDEX PROCCESS ADDRESS ADDRESSES...`

## Requirements
go >= 1.16.3
