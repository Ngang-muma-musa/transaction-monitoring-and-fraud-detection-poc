# Real-Time Transaction Monitoring & Fraud Detection Platform

A distributed system built with **Go**, **Redis**, **WebSockets**, and **message queues** to simulate how real fintech platforms detect fraud in real time.

This project demonstrates:

- Real-time transaction ingestion  
- Asynchronous distributed processing  
- Fraud detection (rule-based + scoring)  
- WebSocket live dashboard updates  
- Event-driven architecture  
- Persistent data and audit logs  

---

## 🚀 Features

### 🔹 Real-Time Transaction Processing
Incoming transactions are validated and streamed through a distributed pipeline using queues.

### 🔹 Fraud Detection Engine
The system evaluates each transaction using:
- configurable rules  
- real-time counters  
- device/IP analysis  
- anomaly detection  
- risk scoring

### 🔹 WebSocket Live Updates
Fraud alerts and risk statuses are broadcast to clients instantly.

### 🔹 Redis for Ultra-Fast State
Used for:
- risk cache  
- fraud flags  
- counters  
- rate limiting  

### 🔹 Event Store & Database
Durable logging of:
- transactions  
- fraud events  
- system activity  
- audit trails  

---

## 🏗️ Architecture Diagram

```mermaid
flowchart LR
    %% STYLES
    classDef svc fill:#1f77b4,stroke:#fff,stroke-width:1,color:#fff
    classDef queue fill:#ff7f0e,stroke:#fff,stroke-width:1,color:#fff
    classDef db fill:#2ca02c,stroke:#fff,stroke-width:1,color:#fff
    classDef cache fill:#9467bd,stroke:#fff,stroke-width:1,color:#fff
    classDef ws fill:#d62728,stroke:#fff,stroke-width:1,color:#fff

    %% CLIENT
    UserClient["User App / Dashboard<br/>WebSocket Stream"]:::ws

    %% API GATEWAY
    subgraph Gateway["API Layer"]
        API["Transaction API Service"]:::svc
        WS["WebSocket Gateway"]:::ws
    end

    %% QUEUE
    Queue[/"Transaction Queue<br/>NATS / Kafka / RabbitMQ"/]:::queue

    %% WORKERS
    subgraph Workers["Distributed Workers"]
        TxWorker["Transaction Processor"]:::svc
        FraudEngine["Fraud Engine<br/>Rules + Scoring"]:::svc
        Notifier["Alert Notification Service"]:::svc
    end

    %% STORAGE
    subgraph Storage["Data Layer"]
        DB["Transaction Database"]:::db
        Cache["Risk Cache<br/>Redis"]:::cache
        EventStore["Event Store<br/>Audit Log"]:::db
    end

    %% ANALYTICS
    subgraph Analytics["Monitoring & Metrics"]
        Metrics["Metrics Collector"]:::svc
        Dashboard["Risk Dashboard<br/>Live Updates"]:::ws
    end

    %% FLOWS
    UserClient -->|REST: Submit Tx| API
    API -->|Publish| Queue

    Queue --> TxWorker
    TxWorker -->|Save Tx| DB
    TxWorker -->|Enriched Event| FraudEngine

    FraudEngine -->|Write Risk Score| DB
    FraudEngine -->|Set Flags| Cache
    FraudEngine -->|Fraud Event| Notifier
    FraudEngine --> EventStore

    Notifier --> WS
    WS -->|Real-Time Updates| UserClient

    DB --> Metrics
    Cache --> Metrics
    EventStore --> Metrics
    Metrics --> Dashboard
    Dashboard --> UserClient

```

---

## 📁 Project Structure

