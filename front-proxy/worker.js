/**
 * Cloudflare Worker that serves as:
 * 1. A personal web page
 * 2. A configurable reverse proxy for services
 */

// Import service configuration
// This will be replaced by wrangler with the actual config from GitHub
const CONFIG = {
  "services": {
    "proxmox": {
      "path": "/proxmox",
      "target": "https://home.stanodn.org:8006",
      "description": "Proxmox VE"
    },
    "vps": {
      "path": "/vps",
      "target": "https://vps.stanodn.org",
      "description": "VPS Management"
    },
    "s3": {
      "path": "/s3",
      "target": "https://home.stanodn.org:9000",
      "description": "S3 Storage"
    }
  },
  "site": {
    "title": "Stanodn Infrastructure",
    "description": "Personal web services dashboard",
    "author": "Stanislav Odnorog",
    "primaryColor": "#3498db",
    "backgroundColor": "#f8f9fa",
    "logo": "https://avatars.githubusercontent.com/u/your-github-username",
    "github": "https://github.com/your-github-username",
    "linkedin": "https://linkedin.com/in/your-linkedin-username",
    "hh": "https://hh.ru/resume/your-resume-id"
  }
};

// HTML template for homepage
function generateHomepage() {
  const services = CONFIG.services;
  const site = CONFIG.site;
  
  let serviceCardsHtml = '';
  for (const [id, service] of Object.entries(services)) {
    serviceCardsHtml += `
      <a href="${service.path}" class="service-card">
        <h3>${id}</h3>
        <p>${service.description}</p>
        <span class="service-link">${service.path}</span>
      </a>
    `;
  }
  
  // Generate social links HTML if they exist in the config
  let socialLinksHtml = '';
  if (site.github || site.linkedin || site.hh) {
    socialLinksHtml = '<div class="social-links">';
    
    if (site.github) {
      socialLinksHtml += `<a href="${site.github}" target="_blank" class="social-link">GitHub</a>`;
    }
    
    if (site.linkedin) {
      socialLinksHtml += `<a href="${site.linkedin}" target="_blank" class="social-link">LinkedIn</a>`;
    }
    
    if (site.hh) {
      socialLinksHtml += `<a href="${site.hh}" target="_blank" class="social-link">HeadHunter</a>`;
    }
    
    socialLinksHtml += '</div>';
  }
  
  return `
  <!DOCTYPE html>
  <html>
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${site.title}</title>
    <style>
      :root {
        --primary-color: ${site.primaryColor || '#3498db'};
        --background-color: ${site.backgroundColor || '#f8f9fa'};
      }
      * { box-sizing: border-box; }
      body {
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
        line-height: 1.6;
        color: #333;
        background-color: var(--background-color);
        margin: 0;
        padding: 0;
      }
      .container {
        max-width: 1200px;
        margin: 0 auto;
        padding: 2rem;
      }
      header {
        margin-bottom: 2rem;
        text-align: center;
        display: flex;
        flex-direction: column;
        align-items: center;
      }
      .profile-pic {
        width: 120px;
        height: 120px;
        border-radius: 50%;
        margin-bottom: 1rem;
        border: 3px solid var(--primary-color);
      }
      h1 {
        color: var(--primary-color);
        margin-bottom: 0.5rem;
      }
      .subtitle {
        color: #666;
        font-size: 1.1rem;
      }
      .service-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 1.5rem;
        margin-top: 2rem;
      }
      .service-card {
        background-color: #fff;
        border-radius: 8px;
        box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        padding: 1.5rem;
        text-decoration: none;
        color: inherit;
        transition: transform 0.2s ease, box-shadow 0.2s ease;
      }
      .service-card:hover {
        transform: translateY(-5px);
        box-shadow: 0 10px 20px rgba(0,0,0,0.1);
      }
      .service-card h3 {
        color: var(--primary-color);
        margin-top: 0;
        margin-bottom: 0.75rem;
        font-size: 1.4rem;
      }
      .service-card p {
        color: #666;
        margin-bottom: 1rem;
      }
      .service-link {
        display: inline-block;
        color: var(--primary-color);
        font-weight: 500;
      }
      footer {
        margin-top: 3rem;
        text-align: center;
        color: #666;
        padding: 1rem;
        font-size: 0.9rem;
      }
      .social-links {
        display: flex;
        justify-content: center;
        gap: 1.5rem;
        margin: 1rem 0;
      }
      .social-link {
        color: var(--primary-color);
        text-decoration: none;
        font-weight: 500;
        transition: color 0.2s;
      }
      .social-link:hover {
        text-decoration: underline;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <header>
        ${site.logo ? `<img src="${site.logo}" alt="${site.author}" class="profile-pic" />` : ''}
        <h1>${site.title}</h1>
        <div class="subtitle">${site.description}</div>
      </header>
      
      <div class="service-grid">
        ${serviceCardsHtml}
      </div>
      
      <footer>
        ${socialLinksHtml}
        <p>&copy; ${new Date().getFullYear()} ${site.author}. All rights reserved.</p>
      </footer>
    </div>
  </body>
  </html>
  `;
}

/**
 * Handle all requests
 */
async function handleRequest(request) {
  const url = new URL(request.url);
  const path = url.pathname;
  
  // Serve homepage at the root
  if (path === "/" || path === "") {
    return new Response(generateHomepage(), {
      headers: { "Content-Type": "text/html" }
    });
  }
  
  // Find service that matches the path
  let targetService = null;
  let remainingPath = "";
  
  for (const [id, service] of Object.entries(CONFIG.services)) {
    if (path.startsWith(service.path)) {
      targetService = service;
      remainingPath = path.substring(service.path.length) || "/";
      break;
    }
  }
  
  // If no service found, return 404
  if (!targetService) {
    return new Response("Service not found", { status: 404 });
  }
  
  // Create target URL by combining service target with remaining path
  const targetUrl = new URL(remainingPath, targetService.target);
  targetUrl.search = url.search; // Preserve query parameters
  
  // Clone the request
  let newHeaders = new Headers(request.headers);
  
  // Create new request for the target
  const newRequest = new Request(targetUrl.toString(), {
    method: request.method,
    headers: newHeaders,
    body: request.body,
    redirect: "follow"
  });
  
  try {
    // Fetch from target and return response
    const response = await fetch(newRequest);
    
    // Clone the response to modify headers if needed
    return new Response(response.body, response);
  } catch (err) {
    return new Response(`Error proxying to service: ${err.message}`, { 
      status: 500,
      headers: { "Content-Type": "text/plain" }
    });
  }
}

// Register event listener for fetch
addEventListener("fetch", event => {
  event.respondWith(handleRequest(event.request));
}); 