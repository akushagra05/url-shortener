import { useState } from 'react'
import URLShortener from './components/URLShortener'
import URLList from './components/URLList'
import Analytics from './components/Analytics'
import Header from './components/Header'
import Footer from './components/Footer'

function App() {
  const [selectedURL, setSelectedURL] = useState(null)
  const [refreshList, setRefreshList] = useState(0)

  const handleURLCreated = () => {
    setRefreshList(prev => prev + 1)
  }

  const handleViewAnalytics = (url) => {
    setSelectedURL(url)
  }

  const handleCloseAnalytics = () => {
    setSelectedURL(null)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <Header />
      
      <main className="container mx-auto px-4 py-8 max-w-6xl">
        {/* URL Shortener Section */}
        <div className="mb-8">
          <URLShortener onURLCreated={handleURLCreated} />
        </div>

        {/* Analytics Modal */}
        {selectedURL && (
          <Analytics 
            shortCode={selectedURL.short_code} 
            onClose={handleCloseAnalytics}
          />
        )}

        {/* URL List Section */}
        <div>
          <URLList 
            refresh={refreshList}
            onViewAnalytics={handleViewAnalytics}
          />
        </div>
      </main>

      <Footer />
    </div>
  )
}

export default App
