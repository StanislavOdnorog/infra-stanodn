# Front Service

A Cloudflare Worker that serves as both:
1. A personal web page with service directory
2. A reverse proxy for various backend services

## Features

- Simple configuration via JSON
- Automatic homepage generation with service cards
- Proxy requests to internal services
- Easy to deploy with GitHub Actions and Cloudflare

## How to Use

### Local Development

1. Install dependencies:
   ```bash
   npm install
   ```

2. Run locally:
   ```bash
   npm run dev
   ```

3. Make changes to `config.json` to update services and site information.

### Deployment

#### Manual Deployment

1. Build the worker:
   ```bash
   npm run build
   ```

2. Deploy to Cloudflare:
   ```bash
   npm run publish
   ```

#### GitHub Actions Deployment

This project is configured to deploy automatically via GitHub Actions whenever changes are pushed to the main branch.

### Configuration

Edit the `config.json` file to:

1. **Add or modify services**:
   ```json
   "services": {
     "new-service": {
       "path": "/path-to-service",
       "target": "https://service-backend-url",
       "description": "Service Description",
       "icon": "icon-name"
     }
   }
   ```

2. **Change site settings**:
   ```json
   "site": {
     "title": "Your Site Title",
     "description": "Your site description",
     "author": "Your Name",
     "primaryColor": "#your-color",
     "backgroundColor": "#your-bg-color"
   }
   ```

## How It Works

1. When a user visits the root path, they see a generated homepage with all configured services
2. When a user navigates to a service path, the request is proxied to the corresponding backend service
3. The worker handles all request routing and proxying

## Notes for Cloudflare Setup

1. Make sure your Cloudflare account has Workers enabled
2. Set up a custom domain or subdomain to point to this worker
3. For secure proxying to services with self-signed certificates, use `originRequest.noTLSVerify: true` in your config 