### Summary of Requirements and Allowed Technologies

#### 1. **Followers**
   - **Required Features**: Users can follow and unfollow others. Following someone requires sending a follow request (unless the profile is public).
   - **Allowed Technologies**: Sessions and cookies for authentication, SQLite for storing follower data.

#### 2. **Profile**
   - **Required Features**: User profiles must display their information, posts, followers, and following. Users can switch between public and private profiles.
   - **Allowed Technologies**: SQLite for storing profile data, image handling for avatars.

#### 3. **Posts**
   - **Required Features**: Users can create posts with images or GIFs. Posts can have privacy settings: public, private, or restricted to certain followers.
   - **Allowed Technologies**: Image handling (JPEG, PNG, GIF), SQLite for storing post data.

#### 4. **Groups**
   - **Required Features**: Users can create groups, invite others, and accept or reject membership requests. Members can post and comment within the group, and create group events.
   - **Allowed Technologies**: SQLite for managing groups, posts, and events.

#### 5. **Chat**
   - **Required Features**: Users can send private messages to users they follow or are followed by. Group chat should also be available for group members.
   - **Allowed Technologies**: **WebSockets** for real-time messaging, **Gorilla WebSocket** for implementation.

#### 6. **Notifications**
   - **Required Features**: Users should receive notifications for:
     - Follow requests
     - Group invitations
     - Group membership requests
     - New group events
   - **Allowed Technologies**: SQLite for notification storage, real-time updates with WebSockets for certain notifications.

#### 7. **Frontend**
   - **Required Features**: Responsive, performant frontend for users to interact with all features (follow, posts, chat, etc.).
   - **Allowed Technologies**: A **JavaScript framework** is required, such as **Next.js, Vue.js, Svelte,** or **Mithril**.

#### 8. **Backend**
   - **Required Features**: A server to handle user requests, manage sessions, and interact with the database. Must handle:
     - User authentication (sessions and cookies)
     - Image uploads (JPEG, PNG, GIF)
     - Real-time chat via WebSockets
     - Group and event handling
   - **Allowed Technologies**: **Caddy** for serving apps, **Go** for the backend logic, **Gorilla WebSocket** for WebSockets, **SQLite** for database.

#### 9. **Database**
   - **Required Features**: Manage user accounts, posts, followers, groups, and chats. Support migrations for schema updates.
   - **Allowed Technologies**: **SQLite**, **golang-migrate** for migrations.

#### 10. **Authentication**
   - **Required Features**: Users must register with email, password, and personal details. Logins should persist via sessions and cookies.
   - **Allowed Technologies**: Custom session and cookie handling or any Go packages that assist with session management.

#### 11. **Migrations**
   - **Required Features**: Database migrations to create necessary tables for users, posts, groups, etc.
   - **Allowed Technologies**: **golang-migrate** or similar migration tools.

#### 12. **Docker**
   - **Required Features**: Containerization of both frontend and backend.
   - **Allowed Technologies**: Docker for container management, with separate images for frontend and backend.

---

### Allowed Packages Summary:
- **Gorilla WebSocket** for real-time chat and notifications.
- **Golang-migrate, sql-migrate, Boostport/migration** for database migrations.
- **SQLite3** for database management.
- **bcrypt** for password hashing and security.
- **UUID** for unique identifiers.

