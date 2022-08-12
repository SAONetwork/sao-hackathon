# go-mcs-sdk

[![Made by SaoNetwork](https://img.shields.io/badge/made%20by-SaoNetwork-green.svg)](https://sao.network/)
[![Chat on discord](https://img.shields.io/badge/join%20-discord-brightgreen.svg)](https://discord.com/invite/q58XsnQqQF)

# Table of Contents <!-- omit in toc -->

- [Introduction](#introduction)
    - [Prerequisites](#prerequisites)
- [MCS API](#mcs-api)
- [Usage](#usage)
    - [Installation](#installation)
    - [Getting Started](#getting-started)
- [Contributing](#contributing)

# Introduction

A golang software development kit for the Multi-Chain Storage (MCS) https://mcs.filswan.com service. It provides a convenient interface for working with the MCS API from a web browser or Node.js. This SDK has the following functionalities:

- **POST**    upload file to Filswan IPFS gate way
- **POST**    make payment to swan filecoin storage gate way
- **POST**    mint asset as NFT(TODO)
- **GET**       list of files uploaded(TODO)
- **GET**       files by cid(TODO)
- **GET**       status from filecoin(TODO)

## Prerequisites

[Go](https://golang.org/doc/install) - Minimum version: 1.17  
Polygon Mumbai Testnet Wallet - [Metamask Tutorial](https://docs.filswan.com/getting-started/beginner-walkthrough/public-testnet/setup-metamask) \
Polygon Mumbai Testnet RPC - [Signup via Alchemy](https://www.alchemy.com/)

You will also need Testnet USDC and MATIC balance to use this SDK. [Swan Faucet Tutorial](https://docs.filswan.com/development-resource/swan-token-contract/acquire-testnet-usdc-and-matic-tokens)

# MCS API

For more information about the API usage, check out the [MCS API documentation](https://docs.filswan.com/development-resource/mcp-api-1).

# Usage

Instructions for developers working with MCS SDK and API.

## Installation
```
go get https://github.com/SaoNetowrk/go-mcs-sdk
```

## Getting Started
Example of uploading a single file using the MCS SDK.

```golang

```

# Contributing

Feel free to join in and discuss. Suggestions are welcome! [Open an issue](https://github.com/SaoNetwork/go-mcs-sdk/issues) or [Join the Discord](https://discord.com/invite/q58XsnQqQF)!