import { useState, useEffect } from 'react'
import axios from 'axios'

const URLList = ({ refresh, onViewAnalytics }) => {
  const [urls, setUrls] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Note: This is a simplified version. In production, you'd need:
  // - Pagination
  // - User authentication to fetch user's URLs
  // - Backend endpoint to list URLs
  // For demo purposes, we'll store created URLs in localStorage

  useEffect(() => {
    loadURLs()
  }, [refresh])

  const loadURLs = () => {
    try {
      const storedURLs = localStorage.getItem('shortURLs')
      if (storedURLs) {
        setUrls(JSON.parse(storedURLs))
      }
    } catch (err) {
      console.error('Failed to load URLs:', err)
    }
  }

  const handleDelete = async (shortCode) => {
    if (!confirm('Are you sure you want to delete this URL?')) {
      return
    }

    try {
      await axios.delete(`/api/v1/url/${shortCode}`)
      
      // Remove from localStorage
      const updatedURLs = urls.filter(url => url.short_code !== shortCode)
      localStorage.setItem('shortURLs', JSON.stringify(updatedURLs))
      setUrls(updatedURLs)
    } catch (err) {
      alert('Failed to delete URL: ' + (err.response?.data?.error?.message || err.message))
    }
  }

  const handleCopy = (shortURL) => {
    navigator.clipboard.writeText(shortURL)
    alert('Copied to clipboard!')
  }

  if (urls.length === 0) {
    return (
      <div className="card text-center py-12">
        <svg className="w-16 h-16 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
        </svg>
        <h3 className="text-lg font-medium text-gray-900 mb-2">No URLs yet</h3>
        <p className="text-gray-600">Create your first short URL above to get started</p>
      </div>
    )
  }

  return (
    <div className="card">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Your Short URLs</h2>
      
      <div className="space-y-4">
        {urls.map((url) => (
          <div key={url.short_code} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition duration-200">
            <div className="flex items-start justify-between">
              <div className="flex-1 min-w-0">
                {/* Short URL */}
                <div className="flex items-center space-x-2 mb-2">
                  <a
                    href={url.short_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary-600 hover:text-primary-700 font-medium text-lg truncate"
                  >
                    {url.short_url}
                  </a>
                  <button
                    onClick={() => handleCopy(url.short_url)}
                    className="text-gray-400 hover:text-gray-600"
                    title="Copy to clipboard"
                  >
                    <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                      <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                    </svg>
                  </button>
                </div>

                {/* Original URL */}
                <p className="text-sm text-gray-600 truncate mb-2">
                  {url.original_url}
                </p>

                {/* Metadata */}
                <div className="flex items-center space-x-4 text-xs text-gray-500">
                  <span>
                    Created: {new Date(url.created_at).toLocaleDateString()}
                  </span>
                  {url.expires_at && (
                    <span className="flex items-center">
                      <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
                      </svg>
                      Expires: {new Date(url.expires_at).toLocaleDateString()}
                    </span>
                  )}
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center space-x-2 ml-4">
                <button
                  onClick={() => onViewAnalytics(url)}
                  className="px-3 py-1 text-sm bg-blue-50 text-blue-700 rounded hover:bg-blue-100 transition duration-200"
                >
                  Analytics
                </button>
                <button
                  onClick={() => handleDelete(url.short_code)}
                  className="px-3 py-1 text-sm bg-red-50 text-red-700 rounded hover:bg-red-100 transition duration-200"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default URLList
