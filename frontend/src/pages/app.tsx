import Shell from '../components/menu/shell'
import Hero from '../components/components/home/hero'
import Features from '../components/components/home/features'
import '../assets/css/home.css'

function App() {
  return (
    <Shell>
      <main className="home">
        <Hero />
        <Features />
      </main>
    </Shell>
  )
}
export default App