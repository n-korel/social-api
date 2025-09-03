import { useNavigate, useParams } from 'react-router-dom'
import { useState } from 'react'
import { API_URL } from './App'

function ConfirmPage() {
  const { token = '' } = useParams()
  const redirect = useNavigate()
  const [loading, setLoading] = useState(false)

  const handleConfirm = async () => {
    setLoading(true)
    try {
      const response = await fetch(`${API_URL}/users/activate/${token}`, {
        method: "PUT"
      })

      if (response.ok) {
        redirect("/")
      } else {
        alert("Failed to confirm token")
      }
    } catch (error) {
      alert("Something went wrong")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="shadow-lg rounded-2xl p-8 max-w-md w-full text-center">
        <h1 className="text-2xl font-bold text-white mb-4">Confirm Your Account</h1>
        <p className="text-white mb-6">
          Click the button below to activate your account.
        </p>
        <button 
          onClick={handleConfirm} 
          disabled={loading}
          className={`px-6 py-2 font-semibold rounded-lg transition ${
            loading 
              ? "bg-blue-800 text-white cursor-not-allowed" 
              : "bg-blue-900 text-white hover:bg-blue-950"
          }`}
        >
          {loading ? (
            <span className="flex items-center justify-center gap-2">
              <svg
                className="animate-spin h-5 w-5 text-white"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
              </svg>
              Confirming...
            </span>
          ) : (
            "Confirm Account"
          )}
        </button>
      </div>
    </div>
  )
}

export default ConfirmPage
