FROM node:18-alpine

LABEL Description="Social-Network Frontend"
LABEL Version="1.0"

WORKDIR /app

# Copy package files first for better layer caching
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application
COPY . .

# Expose the application port
EXPOSE 5173

# Start the development server with host flag to allow external connections
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]