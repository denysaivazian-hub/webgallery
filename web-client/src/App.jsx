import { useEffect } from "react"

export default function App() {
  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch('http://localhost:8090/')
      const respBody = await res.json()
      console.log(respBody)
    }
    fetchData()

    return () => {}
  }, [])
  return (
    <main>
      <h1>Webgallery App</h1>
    </main>
  )
}
