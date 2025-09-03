# Resource Server API Flow Diagrams

## 1. System Architecture Overview

```mermaid
graph TB
    Client[Client Application]
    API[Resource Server API]
    DB[(PostgreSQL Database)]
    
    subgraph "Storage Providers"
        CDN[CDN Provider]
        GCS[Google Cloud Storage]
        R2[Cloudflare R2]
    end
    
    subgraph "API Handlers"
        HealthCheck[Health Check]
        ResourceDef[Resource Definition]
        ProviderHandler[Provider Info]
        FileOps[File Operations]
        MultipartHandler[Multipart Upload]
        AchievementHandler[Achievement]
    end
    
    Client --> API
    API --> HealthCheck
    API --> ResourceDef
    API --> ProviderHandler
    API --> FileOps
    API --> MultipartHandler
    API --> AchievementHandler
    
    API --> DB
    API --> CDN
    API --> GCS
    API --> R2
    
    classDef clientClass fill:#e1f5fe
    classDef apiClass fill:#f3e5f5
    classDef handlerClass fill:#e8f5e8
    classDef storageClass fill:#fff3e0
    classDef dbClass fill:#fce4ec
    
    class Client clientClass
    class API apiClass
    class HealthCheck,ResourceDef,ProviderHandler,FileOps,MultipartHandler,AchievementHandler handlerClass
    class CDN,GCS,R2 storageClass
    class DB dbClass
```

## 2. Health Check & Core APIs Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant DB
    
    %% Health Check
    Client->>API: GET /api/v1/health
    API->>DB: Check connection
    DB-->>API: Connection OK
    API-->>Client: 200 OK + health status
    
    %% Resource Definitions
    Client->>API: GET /api/v1/resources/definitions
    API-->>Client: Achievement & workout definitions
    
    Client->>API: GET /api/v1/resources/definitions/achievement
    API-->>Client: Achievement definition details
    
    %% Provider Information
    Client->>API: GET /api/v1/resources/providers
    API-->>Client: CDN, GCS, R2 capabilities
    
    Client->>API: GET /api/v1/resources/providers/r2
    API-->>Client: R2 specific capabilities
```

## 3. File Operations Flow

```mermaid
graph TD
    Client[Client]
    FileOps[File Operations Handler]
    
    subgraph "Storage Providers"
        CDN[CDN]
        GCS[GCS]
        R2[R2]
    end
    
    %% File Listing
    Client -->|1. GET /resources/:provider/:definition| FileOps
    FileOps -->|2. List files| CDN
    FileOps -->|2. List files| GCS
    FileOps -->|2. List files| R2
    CDN -->|3. File list| FileOps
    GCS -->|3. File list| FileOps
    R2 -->|3. File list| FileOps
    FileOps -->|4. Paginated results| Client
    
    %% Upload URL Generation
    Client -->|5. POST /upload| FileOps
    FileOps -->|6. Generate signed URL| CDN
    FileOps -->|6. Generate signed URL| GCS
    FileOps -->|6. Generate signed URL| R2
    CDN -->|7. Signed URL| FileOps
    GCS -->|7. Signed URL| FileOps
    R2 -->|7. Signed URL| FileOps
    FileOps -->|8. Upload URL| Client
    
    %% Direct Upload
    Client -.->|9. Upload file| CDN
    Client -.->|9. Upload file| GCS
    Client -.->|9. Upload file| R2
    
    classDef clientClass fill:#e1f5fe
    classDef handlerClass fill:#e8f5e8
    classDef storageClass fill:#fff3e0
    
    class Client clientClass
    class FileOps handlerClass
    class CDN,GCS,R2 storageClass
```

## 4. Download & Metadata Operations

```mermaid
sequenceDiagram
    participant Client
    participant FileOps as File Operations
    participant Storage as Storage Provider
    
    %% Download Flow
    Client->>FileOps: POST /resources/:provider/*/download
    FileOps->>Storage: Generate download URL
    Storage-->>FileOps: Signed download URL
    FileOps-->>Client: Download URL + headers
    Client->>Storage: Download file (direct)
    Storage-->>Client: File content
    
    %% Metadata Operations
    Client->>FileOps: GET /resources/:provider/*/metadata
    FileOps->>Storage: Get file metadata
    Storage-->>FileOps: Metadata (size, type, etc.)
    FileOps-->>Client: Metadata response
    
    Client->>FileOps: PUT /resources/:provider/*/metadata
    FileOps->>Storage: Update metadata
    Storage-->>FileOps: Update confirmation
    FileOps-->>Client: Success response
    
    %% Delete Operation
    Client->>FileOps: DELETE /resources/:provider/*
    FileOps->>Storage: Delete file
    Storage-->>FileOps: Deletion confirmed
    FileOps-->>Client: Success response
```

## 5. Multipart Upload Flow

```mermaid
graph TD
    Client[Client]
    MultipartHandler[Multipart Handler]
    DB[(Database)]
    Storage[Storage Provider]
    
    %% Initialize
    Client -->|1. POST /multipart/init| MultipartHandler
    MultipartHandler -->|2. Store metadata| DB
    MultipartHandler -->|3. Initialize multipart| Storage
    Storage -->|4. Upload ID| MultipartHandler
    MultipartHandler -->|5. Upload ID + constraints| Client
    
    %% Get Part URLs
    Client -->|6. POST /multipart/urls| MultipartHandler
    MultipartHandler -->|7. Get metadata| DB
    MultipartHandler -->|8. Generate part URLs| Storage
    Storage -->|9. Part URLs| MultipartHandler
    MultipartHandler -->|10. Part URLs + complete/abort URLs| Client
    
    %% Upload Parts
    Client -.->|11. Upload parts| Storage
    Client -.->|12. Complete multipart| Storage
    
    classDef clientClass fill:#e1f5fe
    classDef handlerClass fill:#e8f5e8
    classDef storageClass fill:#fff3e0
    classDef dbClass fill:#fce4ec
    
    class Client clientClass
    class MultipartHandler handlerClass
    class Storage storageClass
    class DB dbClass
```

## 6. Achievement Management Flow

```mermaid
sequenceDiagram
    participant Client
    participant AchievementHandler as Achievement Handler
    participant DB as Database
    participant Storage as Storage Provider
    
    %% List Achievements
    Client->>AchievementHandler: GET /achievements/
    AchievementHandler->>DB: Query achievements
    DB-->>AchievementHandler: Achievement list
    AchievementHandler-->>Client: Paginated achievements
    
    %% Create Achievement
    Client->>AchievementHandler: POST /achievements/
    AchievementHandler->>DB: Create achievement
    AchievementHandler->>Storage: Generate upload URL
    Storage-->>AchievementHandler: Upload URL
    DB-->>AchievementHandler: Achievement created
    AchievementHandler-->>Client: Achievement + upload URL
    
    %% Get Achievement
    Client->>AchievementHandler: GET /achievements/:id
    AchievementHandler->>DB: Query by ID
    DB-->>AchievementHandler: Achievement details
    AchievementHandler-->>Client: Achievement data
    
    %% Update Icon
    Client->>AchievementHandler: PUT /achievements/:id/icon
    AchievementHandler->>DB: Update icon metadata
    AchievementHandler->>Storage: Generate new upload URL
    Storage-->>AchievementHandler: New upload URL
    DB-->>AchievementHandler: Update confirmed
    AchievementHandler-->>Client: New upload URL
```

## 7. Upload Confirmation & Verification Flow

```mermaid
graph TD
    Client[Client]
    AchievementHandler[Achievement Handler]
    DB[(Database)]
    Storage[Storage Provider]
    
    %% Upload Confirmation
    Client -->|1. POST /uploads/:id/confirm| AchievementHandler
    AchievementHandler -->|2. Update upload status| DB
    AchievementHandler -->|3. Verify file exists| Storage
    Storage -->|4. File verification| AchievementHandler
    DB -->|5. Status updated| AchievementHandler
    AchievementHandler -->|6. Confirmation response| Client
    
    %% Multipart URLs for Achievement
    Client -->|7. POST /uploads/:id/multipart| AchievementHandler
    AchievementHandler -->|8. Get upload metadata| DB
    AchievementHandler -->|9. Generate part URLs| Storage
    Storage -->|10. Part URLs| AchievementHandler
    AchievementHandler -->|11. Multipart URLs| Client
    
    classDef clientClass fill:#e1f5fe
    classDef handlerClass fill:#e8f5e8
    classDef storageClass fill:#fff3e0
    classDef dbClass fill:#fce4ec
    
    class Client clientClass
    class AchievementHandler handlerClass
    class Storage storageClass
    class DB dbClass
```

## Flow Summary

### Diagram 1: System Architecture
- Shows high-level system components and relationships
- 6 main handlers connecting client to storage providers and database

### Diagram 2: Health Check & Core APIs
- Simple request/response patterns for system status and configuration
- No complex business logic, just data retrieval

### Diagram 3: File Operations
- Multi-provider file listing and upload URL generation
- Shows both API flow and direct client-to-storage uploads

### Diagram 4: Download & Metadata
- Download URL generation and direct file access
- Metadata CRUD operations across providers

### Diagram 5: Multipart Upload
- Complex workflow for large file uploads
- Database tracking and multi-step URL generation

### Diagram 6: Achievement Management
- CRUD operations with integrated file upload support
- Database persistence with storage provider integration

### Diagram 7: Upload Confirmation
- Post-upload verification and status tracking
- Achievement-specific multipart upload workflows

Each diagram focuses on a specific domain, making them easier to read and understand while covering all the flows from your end-to-end test plan.

---

# Client-Server Communication Only

## 1. System Overview - Client-Server

```mermaid
graph TB
    Client[Client Application]
    Server[Resource Server API]
    
    Client <--> Server
    
    classDef clientClass fill:#e1f5fe
    classDef serverClass fill:#f3e5f5
    
    class Client clientClass
    class Server serverClass
```

## 2. Health Check & Core APIs - Client-Server

```mermaid
sequenceDiagram
    participant Client
    participant Server as Resource Server
    
    Client->>Server: GET /api/v1/health
    Server-->>Client: 200 OK + health status
    
    Client->>Server: GET /api/v1/resources/definitions
    Server-->>Client: Achievement & workout definitions
    
    Client->>Server: GET /api/v1/resources/definitions/achievement
    Server-->>Client: Achievement definition details
    
    Client->>Server: GET /api/v1/resources/providers
    Server-->>Client: CDN, GCS, R2 capabilities
    
    Client->>Server: GET /api/v1/resources/providers/r2
    Server-->>Client: R2 specific capabilities
```

## 3. File Operations - Client-Server

```mermaid
sequenceDiagram
    participant Client
    participant Server as Resource Server
    
    %% File Listing
    Client->>Server: GET /api/v1/resources/:provider/:definition
    Server-->>Client: Paginated file list
    
    %% Upload URL Generation
    Client->>Server: POST /api/v1/resources/:provider/:definition/upload
    Server-->>Client: Signed upload URL + metadata
    
    %% Download URL Generation
    Client->>Server: POST /api/v1/resources/:provider/*/download
    Server-->>Client: Signed download URL
    
    %% Metadata Operations
    Client->>Server: GET /api/v1/resources/:provider/*/metadata
    Server-->>Client: File metadata
    
    Client->>Server: PUT /api/v1/resources/:provider/*/metadata
    Server-->>Client: Update confirmation
    
    %% Delete Operation
    Client->>Server: DELETE /api/v1/resources/:provider/*
    Server-->>Client: Deletion confirmation
```

## 4. Multipart Upload - Client-Server

```mermaid
sequenceDiagram
    participant Client
    participant Server as Resource Server
    
    %% Initialize Multipart Upload
    Client->>Server: POST /api/v1/resources/multipart/init
    Server-->>Client: Upload ID + constraints
    
    %% Get Part URLs
    Client->>Server: POST /api/v1/resources/multipart/urls
    Server-->>Client: Part URLs + complete/abort URLs
```

## 5. Achievement Management - Client-Server

```mermaid
sequenceDiagram
    participant Client
    participant Server as Resource Server
    
    %% List Achievements
    Client->>Server: GET /api/v1/achievements/
    Server-->>Client: Paginated achievements
    
    %% Create Achievement
    Client->>Server: POST /api/v1/achievements/
    Server-->>Client: Achievement + upload URL
    
    %% Get Achievement
    Client->>Server: GET /api/v1/achievements/:id
    Server-->>Client: Achievement details
    
    %% Update Achievement Icon
    Client->>Server: PUT /api/v1/achievements/:id/icon
    Server-->>Client: New upload URL
    
    %% Confirm Upload
    Client->>Server: POST /api/v1/achievements/uploads/:id/confirm
    Server-->>Client: Confirmation response
    
    %% Get Multipart URLs
    Client->>Server: POST /api/v1/achievements/uploads/:id/multipart
    Server-->>Client: Multipart upload URLs
```

## 6. Complete API Endpoints Summary

```mermaid
graph LR
    Client[Client Application]
    
    subgraph "Resource Server Endpoints"
        Health["/api/v1/health"]
        Definitions["/api/v1/resources/definitions"]
        Providers["/api/v1/resources/providers"]
        Files["/api/v1/resources/:provider/:definition"]
        Upload["/api/v1/resources/:provider/:definition/upload"]
        Download["/api/v1/resources/:provider/*/download"]
        Metadata["/api/v1/resources/:provider/*/metadata"]
        Delete["/api/v1/resources/:provider/*"]
        MultipartInit["/api/v1/resources/multipart/init"]
        MultipartUrls["/api/v1/resources/multipart/urls"]
        Achievements["/api/v1/achievements/"]
        AchievementGet["/api/v1/achievements/:id"]
        AchievementIcon["/api/v1/achievements/:id/icon"]
        UploadConfirm["/api/v1/achievements/uploads/:id/confirm"]
        UploadMultipart["/api/v1/achievements/uploads/:id/multipart"]
    end
    
    Client --> Health
    Client --> Definitions
    Client --> Providers
    Client --> Files
    Client --> Upload
    Client --> Download
    Client --> Metadata
    Client --> Delete
    Client --> MultipartInit
    Client --> MultipartUrls
    Client --> Achievements
    Client --> AchievementGet
    Client --> AchievementIcon
    Client --> UploadConfirm
    Client --> UploadMultipart
    
    classDef clientClass fill:#e1f5fe
    classDef endpointClass fill:#e8f5e8
    
    class Client clientClass
    class Health,Definitions,Providers,Files,Upload,Download,Metadata,Delete,MultipartInit,MultipartUrls,Achievements,AchievementGet,AchievementIcon,UploadConfirm,UploadMultipart endpointClass
```