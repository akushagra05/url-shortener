import { useState, useEffect } from 'react'
import axios from 'axios'

const Analytics = ({ shortCode, onClose }) => {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    loadAnalytics()
  }, [shortCode])

  const loadAnalytics = async () => {
    try {
      setLoading(true)
      const response = await axios.get(`/api/v1/url/${shortCode}/analytics?days=30`)
      
      if (response.data.success) {
        setData(response.data.data)
      }
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to load analytics')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
          <h2 className="text-2xl font-bold text-gray-900">URL Analytics</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition duration-200"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          {loading && (
            <div className="text-center py-12">
              <svg className="animate-spin h-12 w-12 mx-auto text-primary-600" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <p className="mt-4 text-gray-600">Loading analytics...</p>
            </div>
          )}

          {error && (
            <div className="text-center py-12">
              <svg className="w-12 h-12 mx-auto text-red-600 mb-4" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <p className="text-red-800 font-medium">{error}</p>
            </div>
          )}

          {data && !loading && (
            <div className="space-y-6">
              {/* Total Clicks */}
              <div className="bg-gradient-to-r from-primary-500 to-primary-600 rounded-lg p-6 text-white">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-primary-100 text-sm font-medium">Total Clicks</p>
                    <p className="text-4xl font-bold mt-2">{data.total_clicks.toLocaleString()}</p>
                  </div>
                  <svg className="w-16 h-16 text-primary-200" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
                  </svg>
                </div>
              </div>

              {/* Clicks by Date */}
              {data.analytics.clicks_by_date && data.analytics.clicks_by_date.length > 0 && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Clicks Over Time (Last 30 Days)</h3>
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="space-y-2">
                      {data.analytics.clicks_by_date.slice(0, 10).map((item, index) => (
                        <div key={index} className="flex items-center">
                          <span className="text-sm text-gray-600 w-24">{item.date}</span>
                          <div className="flex-1 ml-4">
                            <div className="bg-primary-200 rounded-full h-6 relative overflow-hidden">
                              <div
                                className="bg-primary-600 h-full rounded-full flex items-center justify-end pr-2"
                                style={{
                                  width: `${(item.count / Math.max(...data.analytics.clicks_by_date.map(d => d.count))) * 100}%`
                                }}
                              >
                                <span className="text-xs text-white font-medium">{item.count}</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Device Distribution */}
              {data.analytics.devices && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Device Types</h3>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="bg-blue-50 rounded-lg p-4 text-center">
                      <svg className="w-8 h-8 mx-auto text-blue-600 mb-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M7 2a2 2 0 00-2 2v12a2 2 0 002 2h6a2 2 0 002-2V4a2 2 0 00-2-2H7zm3 14a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
                      </svg>
                      <p className="text-2xl font-bold text-blue-900">{data.analytics.devices.mobile}</p>
                      <p className="text-sm text-blue-700">Mobile</p>
                    </div>
                    <div className="bg-green-50 rounded-lg p-4 text-center">
                      <svg className="w-8 h-8 mx-auto text-green-600 mb-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M3 5a2 2 0 012-2h10a2 2 0 012 2v8a2 2 0 01-2 2h-2.22l.123.489.804.804A1 1 0 0113 18H7a1 1 0 01-.707-1.707l.804-.804L7.22 15H5a2 2 0 01-2-2V5zm5.771 7H5V5h10v7H8.771z" clipRule="evenodd" />
                      </svg>
                      <p className="text-2xl font-bold text-green-900">{data.analytics.devices.desktop}</p>
                      <p className="text-sm text-green-700">Desktop</p>
                    </div>
                    <div className="bg-purple-50 rounded-lg p-4 text-center">
                      <svg className="w-8 h-8 mx-auto text-purple-600 mb-2" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z" />
                      </svg>
                      <p className="text-2xl font-bold text-purple-900">{data.analytics.devices.tablet}</p>
                      <p className="text-sm text-purple-700">Tablet</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Top Countries */}
              {data.analytics.top_countries && data.analytics.top_countries.length > 0 && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Top Countries</h3>
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="space-y-3">
                      {data.analytics.top_countries.slice(0, 5).map((item, index) => (
                        <div key={index} className="flex items-center justify-between">
                          <span className="text-sm font-medium text-gray-700">{item.country}</span>
                          <span className="text-sm text-gray-600">{item.count} clicks</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Top Referrers */}
              {data.analytics.top_referrers && data.analytics.top_referrers.length > 0 && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Top Referrers</h3>
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="space-y-3">
                      {data.analytics.top_referrers.slice(0, 5).map((item, index) => (
                        <div key={index} className="flex items-center justify-between">
                          <span className="text-sm font-medium text-gray-700 truncate flex-1">
                            {item.referrer || 'Direct'}
                          </span>
                          <span className="text-sm text-gray-600 ml-4">{item.count} clicks</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* No Data Message */}
              {data.total_clicks === 0 && (
                <div className="text-center py-8 text-gray-500">
                  <p>No analytics data available yet. Share your link to start collecting data!</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Analytics
