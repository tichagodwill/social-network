# Social Network Platform

## Description
A modern social networking platform built with Go and Svelte, featuring real-time communication and robust security measures. This platform implements core social networking features while maintaining high performance and scalability through efficient database design and WebSocket implementations.

---

## Features

### User Management
- **Authentication**
  - Secure session-based authentication
  - JWT token implementation for API requests
  - Password encryption using bcrypt
  - Email verification system
  - Password recovery functionality

- **Profiles**
  - Customizable user profiles with avatars
  - Profile privacy settings (Public/Private)
  - Bio and personal information
  - Activity timeline
  - Profile verification badges
  - Custom username selection

### Social Features
- **Following System**
  - Follow/Unfollow functionality
  - Follow requests for private profiles
  - Follower/Following lists
  - Suggested users based on mutual connections
  - Block user functionality

- **Posts**
  - Text posts with rich text formatting
  - Image upload support (JPEG, PNG, GIF)
  - Privacy levels:
    - Public (visible to all)
    - Almost Private (visible to followers)
    - Private (visible to selected users)
  - Post editing and deletion
  - Share/Repost functionality
  - Save posts for later

- **Interactions**
  - Comments with threading support
  - Reactions (Like, Love, Laugh, etc.)
  - @mentions and notifications
  - Post sharing
  - Comment moderation tools

### Groups
- **Management**
  - Create public/private groups
  - Customize group settings and rules
  - Assign roles (Admin, Moderator, Member)
  - Member management tools
  - Group discovery feature

- **Content**
  - Group-specific posts
  - Event creation and management
  - Shared files and media
  - Pinned announcements
  - Group chat integration

### Real-time Features
- **Chat System**
  - Private one-on-one messaging
  - Group chats
  - Message status (Sent/Delivered/Read)
  - File sharing in chats
  - Online status indicators
  - Typing indicators
  - Message search functionality

- **Notifications**
  - Real-time push notifications
  - Email notifications (configurable)
  - Activity notifications for:
    - New followers
    - Post interactions
    - Group activities
    - Chat messages
    - System announcements

---

## Technical Architecture

### Backend (Go)
- **Framework**: Custom router based on net/http
- **Database**: SQLite3 with efficient indexing
- **Key Packages**:
  - `github.com/gorilla/websocket` for WebSocket connections
  - `github.com/golang-migrate/migrate` for database migrations
  - `github.com/dgrijalva/jwt-go` for JWT authentication
  - `golang.org/x/crypto/bcrypt` for password hashing

### Frontend (Svelte)
- **State Management**: Svelte stores
- **Key Features**:
  - SSR (Server-Side Rendering)
  - Progressive Web App (PWA) support
  - Responsive design using CSS Grid/Flexbox
  - Custom WebSocket client implementation
  - Image lazy loading and optimization

### Database Schema
```sql
-- Key tables structure
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    profile_type TEXT CHECK(profile_type IN ('public', 'private')) DEFAULT 'public',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id),
    content TEXT NOT NULL,
    privacy_level TEXT CHECK(privacy_level IN ('public', 'almost_private', 'private')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Additional tables defined in migrations
```

---

## Installation

### Prerequisites
- Docker v20.10 or higher
- Docker Compose v2.0 or higher
- Go v1.19 or higher (for local development)
- Node.js v16 or higher (for local development)
- SQLite3 v3.37 or higher

### Local Development Setup
1. Clone the repository:
    ```bash
    git clone https://learn.reboot01.com/git/tnji/social-network.git
    cd social-network
    ```

2. Set up environment variables:
    ```bash
    cp .env.example .env
    # Edit .env with your configurations
    ```

3. Install frontend dependencies:
    ```bash
    cd frontend
    npm install
    ```

4. Install backend dependencies:
    ```bash
    cd backend
    go mod download
    ```

5. Run database migrations:
    ```bash
    cd backend
    go run cmd/migrate/main.go up
    ```

### Docker Deployment
1. Build and start containers:
    ```bash
    docker-compose up --build
    ```

2. Access the application:
    - Frontend: `http://localhost:8080`
    - Backend API: `http://localhost:8081`
    - Adminer (Database UI): `http://localhost:8082`

### Configuration Options
- `JWT_SECRET`: Secret key for JWT token generation
- `COOKIE_SECRET`: Secret for session cookie encryption
- `DB_PATH`: SQLite database file location
- `UPLOAD_DIR`: Directory for storing uploaded files
- `MAX_UPLOAD_SIZE`: Maximum file upload size in bytes

---

## API Documentation

### Authentication Endpoints
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/verify/:token` - Email verification

### User Endpoints
- `GET /api/users/:id` - Get user profile
- `PUT /api/users/:id` - Update user profile
- `GET /api/users/:id/followers` - Get user followers
- `POST /api/users/:id/follow` - Follow user

### Post Endpoints
- `GET /api/posts` - Get feed posts
- `POST /api/posts` - Create new post
- `PUT /api/posts/:id` - Update post
- `DELETE /api/posts/:id` - Delete post

### Complete API documentation available at `/api/docs`

---

## Testing

### Backend Tests
```bash
cd backend
go test ./... -v
```

### Frontend Tests
```bash
cd frontend
npm run test
```

### Integration Tests
```bash
docker-compose -f docker-compose.test.yml up --build
```

---

## Security Measures
- CSRF protection
- XSS prevention
- Rate limiting
- Input validation
- Secure password storage
- Session management
- File upload validation

---

## Performance Optimization
- Database indexing
- Cache implementation
- Image optimization
- Lazy loading
- Connection pooling
- WebSocket connection management

---

## Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Coding Standards
- Go: Follow the official Go style guide
- JavaScript: ESLint with Prettier
- Commit messages: Conventional Commits format

---

## Authors
- **Tnji** - *Initial work and maintenance*

---

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

---

## Acknowledgments
- SQLite documentation and community
- Golang-migrate for database migration handling
- Gorilla WebSocket for real-time communication
- Svelte team for the excellent frontend framework
- The Go community for various packages and inspiration
