import { useState } from 'react'
import axios from 'axios'

const URLShortener = ({ onURLCreated }) => {
  const [url, setUrl] = useState('')
  const [customAlias, setCustomAlias] = useState('')
  const [expiresAt, setExpiresAt] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    setResult(null)

    try {
      const payload = {
        url: url,
      }

      if (customAlias.trim()) {
        payload.custom_alias = customAlias.trim()
      }

      if (expiresAt) {
        payload.expires_at = new Date(expiresAt).toISOString()
      }

      const response = await axios.post('/api/v1/shorten', payload)

      if (response.data.success) {
        setResult(response.data.data)
        
        // Save to localStorage for demo purposes
        const storedURLs = JSON.parse(localStorage.getItem('shortURLs') || '[]')
        storedURLs.unshift(response.data.data)
        localStorage.setItem('shortURLs', JSON.stringify(storedURLs))
        
        setUrl('')
        setCustomAlias('')
        setExpiresAt('')
        if (onURLCreated) {
          onURLCreated()
        }
      }
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to create short URL')
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = () => {
    if (result?.short_url) {
      navigator.clipboard.writeText(result.short_url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="card">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Shorten Your URL</h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* URL Input */}
        <div>
          <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-2">
            Enter your long URL *
          </label>
          <input
            id="url"
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/very/long/url"
            className="input-field"
            required
          />
        </div>

        {/* Custom Alias */}
        <div>
          <label htmlFor="alias" className="block text-sm font-medium text-gray-700 mb-2">
            Custom alias (optional, 4-20 characters)
          </label>
          <input
            id="alias"
            type="text"
            value={customAlias}
            onChange={(e) => setCustomAlias(e.target.value)}
            placeholder="my-custom-link"
            className="input-field"
            minLength={4}
            maxLength={20}
            pattern="[a-zA-Z0-9_-]+"
          />
          <p className="text-xs text-gray-500 mt-1">
            Letters, numbers, hyphens, and underscores only
          </p>
        </div>

        {/* Expiration Date */}
        <div>
          <label htmlFor="expires" className="block text-sm font-medium text-gray-700 mb-2">
            Expiration date (optional)
          </label>
          <input
            id="expires"
            type="datetime-local"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            className="input-field"
            min={new Date().toISOString().slice(0, 16)}
          />
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          disabled={loading}
          className="btn-primary w-full text-lg"
        >
          {loading ? (
            <span className="flex items-center justify-center">
              <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Shortening...
            </span>
          ) : (
            'Shorten URL'
          )}
        </button>
      </form>

      {/* Error Message */}
      {error && (
        <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center">
            <svg className="w-5 h-5 text-red-600 mr-2" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
            <p className="text-red-800 font-medium">{error}</p>
          </div>
        </div>
      )}

      {/* Success Result */}
      {result && (
        <div className="mt-6 p-6 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center mb-3">
            <svg className="w-6 h-6 text-green-600 mr-2" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <h3 className="text-lg font-semibold text-green-900">Success! Your short URL is ready</h3>
          </div>
          
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Short URL</label>
              <div className="flex items-center space-x-2">
                <input
                  type="text"
                  value={result.short_url}
                  readOnly
                  className="flex-1 px-4 py-2 bg-white border border-gray-300 rounded-lg font-mono text-primary-600"
                />
                <button
                  onClick={handleCopy}
                  className="btn-secondary whitespace-nowrap"
                >
                  {copied ? (
                    <span className="flex items-center">
                      <svg className="w-5 h-5 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                      Copied!
                    </span>
                  ) : (
                    <span className="flex items-center">
                      <svg className="w-5 h-5 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                        <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                      </svg>
                      Copy
                    </span>
                  )}
                </button>
              </div>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Original URL</label>
              <p className="text-sm text-gray-600 break-all">{result.original_url}</p>
            </div>

            {result.expires_at && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Expires</label>
                <p className="text-sm text-gray-600">
                  {new Date(result.expires_at).toLocaleString()}
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default URLShortener
