# URL Shortener Frontend

Modern React frontend built with Vite, Tailwind CSS, and Axios.

## Features

- 🎨 Clean, modern UI with Tailwind CSS
- ⚡ Fast development with Vite
- 📱 Responsive design (mobile, tablet, desktop)
- 📊 Analytics dashboard with visualizations
- 📋 Copy to clipboard functionality
- 🔄 Real-time URL list updates
- ⏰ Custom alias and expiration support

## Prerequisites

- Node.js 18+ and npm (or yarn/pnpm)

## Installation

```bash
# Install Node.js and npm
# macOS:
brew install node

# Or download from: https://nodejs.org/

# Install dependencies
cd frontend
npm install
```

## Development

```bash
# Start development server
npm run dev

# The app will be available at http://localhost:5173
```

The frontend is configured to proxy API requests to the backend at `http://localhost:8080`.

## Build for Production

```bash
# Build optimized production bundle
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── Header.jsx         # App header with branding
│   │   ├── Footer.jsx         # App footer
│   │   ├── URLShortener.jsx   # Main URL shortening form
│   │   ├── URLList.jsx        # List of created URLs
│   │   └── Analytics.jsx      # Analytics modal
│   ├── App.jsx                # Main app component
│   ├── main.jsx              # App entry point
│   └── index.css             # Global styles + Tailwind
├── index.html                # HTML template
├── vite.config.js           # Vite configuration
├── tailwind.config.js       # Tailwind configuration
└── package.json             # Dependencies
```

## Components

### URLShortener
Main component for creating short URLs with:
- Long URL input
- Custom alias (optional)
- Expiration date (optional)
- Success/error notifications
- Copy to clipboard

### URLList
Displays created URLs with:
- Short URL and original URL
- Created date and expiration
- Analytics button
- Delete functionality
- Copy to clipboard

### Analytics
Modal showing detailed analytics:
- Total clicks
- Clicks over time (chart)
- Device distribution (mobile/desktop/tablet)
- Top countries
- Top referrers

## Configuration

The frontend is configured via `vite.config.js`:

```javascript
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

For production, update the proxy target to your deployed backend URL.

## Deployment

### Vercel (Recommended)

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy
cd frontend
vercel

# Follow the prompts
```

### Netlify

```bash
# Build
npm run build

# Deploy dist/ folder via Netlify dashboard or CLI
```

### Environment Variables

For production deployment, set these environment variables:
- `VITE_API_URL`: Backend API URL (e.g., https://api.yourapp.com)

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Technologies

- **React 18**: UI library
- **Vite 5**: Build tool and dev server
- **Tailwind CSS 3**: Utility-first CSS framework
- **Axios**: HTTP client
- **PostCSS**: CSS processing
- **Autoprefixer**: CSS vendor prefixing

## Development Tips

### Hot Module Replacement (HMR)
Vite provides instant HMR - your changes appear immediately without full page reload.

### Tailwind CSS IntelliSense
Install the Tailwind CSS IntelliSense VS Code extension for autocompletion.

### React DevTools
Install React DevTools browser extension for debugging.

## Troubleshooting

### Port already in use
```bash
# Change port in vite.config.js or use:
npm run dev -- --port 3000
```

### API requests failing
- Ensure backend is running on port 8080
- Check CORS configuration in backend
- Verify proxy settings in vite.config.js

### Build errors
```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
```

## Future Enhancements

- [ ] QR code generation for URLs
- [ ] Bulk URL upload
- [ ] User authentication
- [ ] Dark mode
- [ ] More chart types for analytics
- [ ] Export analytics data
- [ ] Custom domains

## License

MIT License
