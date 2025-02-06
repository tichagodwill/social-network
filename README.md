# Social Network Project

A Facebook-like social network built with Go (backend) and SvelteKit (frontend), featuring real-time communication, group management, and comprehensive user interactions.

## Features

### Authentication
- User registration with required fields:
  - Email, Password, First/Last Name, Date of Birth
  - Optional: Avatar, Nickname, About Me
- Session-based authentication using cookies
- Persistent login state

### Profile Management
- Public/Private profile options
- Profile customization
- Activity feed showing user's posts
- Followers/Following lists
- User information display

### Posts & Comments
- Create posts with privacy settings:
  - Public: Visible to all users
  - Almost Private: Visible to followers only
  - Private: Visible to selected followers
- Support for media attachments (JPEG, PNG, GIF)
- Comment system with media support
- Like/Unlike functionality

### Groups
- Create groups with title and description
- Invite system for group membership
- Request-to-join functionality
- Group posts and comments
- Event creation with RSVP system
- Group chat functionality

### Real-time Features
- Private messaging between connected users
- Group chat rooms
- Real-time notifications
- Emoji support in messages
- WebSocket implementation for instant updates

### Notifications
- Follow requests for private profiles
- Group invitations
- Join requests for group creators
- Event creation notifications
- Message notifications

## Technical Stack

### Frontend (SvelteKit)
- TypeScript for type safety
- Tailwind CSS for styling
- Flowbite components
- WebSocket client implementation
- Responsive design

### Backend (Go)
- Custom web server
- SQLite database
- Migration system
- WebSocket server
- Session management
- File upload handling

### Database
- SQLite with migrations
- Structured schema for:
  - Users
  - Posts
  - Comments
  - Groups
  - Messages
  - Notifications
  - Followers

## Project Structure

```
├── client/                 # Frontend application
│   ├── src/
│   │   ├── lib/           # Shared components and utilities
│   │   ├── routes/        # SvelteKit routes
│   │   └── app.html       # Main HTML template
│   └── static/            # Static assets
│
├── server/                 # Backend application
│   ├── api/               # API handlers
│   ├── models/            # Data models
│   ├── pkg/
│   │   ├── db/           # Database operations
│   │   │   ├── migrations/
│   │   │   └── sqlite/   # SQLite implementation
│   │   └── util/         # Utilities
│   └── main.go           # Entry point
```

## Getting Started

### Prerequisites
- Docker and Docker Compose
- Go 1.22+
- Node.js 18+

### Development Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd social-network
```

2. Start the backend:
```bash
cd server
go mod download
go run main.go
```

3. Start the frontend:
```bash
cd client
npm install
npm run dev
```

### Docker Setup

Build and run using Docker Compose:
```bash
docker-compose up --build
```

## API Documentation

[To be included later]

## Database Schema

[To be included later]

## Contributing

[To be included later]

## License

[To be included later]
