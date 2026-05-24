import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

export const Route = createFileRoute('/quantum')({
  component: QuantumPage,
})

interface Qubit { id: number; gate: string; angle: number }

function QuantumPage() {
  const { t } = useT()
  const [qubits, setQubits] = useState<Qubit[]>([{ id: 0, gate: 'H', angle: 0 }])
  const [result, setResult] = useState<string | null>(null)
  const [running, setRunning] = useState(false)

  const addQubit = () => setQubits(p => [...p, { id: p.length, gate: 'H', angle: 0 }])
  const removeQubit = (id: number) => setQubits(p => p.filter(q => q.id !== id))

  const simulate = async () => {
    setRunning(true)
    setResult(null)
    // Simulated execution
    await new Promise(r => setTimeout(r, 1500))
    const states = ['|0⟩', '|1⟩', '|+⟩', '|-⟩', '|i⟩', '|-i⟩']
    setResult(states[Math.floor(Math.random() * states.length)])
    setRunning(false)
  }

  return (
    <div className="min-h-screen bg-background"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="qc-wrap qc-section-pad-sm">
        <div className="qc-fade-up mb-10">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-amber-200 bg-amber-50 text-amber-700 text-xs font-semibold tracking-wide mb-5">
            <span className="w-2 h-2 rounded-full bg-amber-500" />
            {t('Quantum Computing')}
          </div>
          <h1 className="qc-title-hero font-bold tracking-tight text-foreground">
            {t('Quantum Playground')}
          </h1>
          <p className="qc-text-body qc-readable-width text-muted-foreground/70 mt-2 leading-relaxed">
            {t('Run quantum circuits on simulated qubits. Experiment with gates and measure outcomes.')}
          </p>
        </div>

        <div className="flex flex-col lg:flex-row gap-8">
          {/* Circuit builder */}
          <div className="flex-1">
            <div className="qc-fade-up rounded-2xl bg-white/70 backdrop-blur-sm border border-border/10 p-6">
              <div className="flex items-center justify-between mb-5">
                <h3 className="text-sm font-bold tracking-tight">{t('Circuit')} ({qubits.length} {t('qubits')})</h3>
                <button onClick={addQubit}
                  className="px-3 py-1.5 rounded-xl text-xs font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 hover:shadow-md transition-all">
                  + {t('Add')}
                </button>
              </div>
              <div className="space-y-3">
                {qubits.map((q, i) => (
                  <div key={q.id}
                    className="qc-fade-up flex items-center gap-3 px-4 py-3 rounded-xl bg-white/80 border border-border/10"
                    style={{ animationDelay: `${i * 0.05}s` }}>
                    <span className="text-xs font-semibold text-muted-foreground/50 w-8">q{q.id}</span>
                    <select value={q.gate} onChange={e => setQubits(p => p.map(x => x.id === q.id ? { ...x, gate: e.target.value } : x))}
                      className="flex-1 h-9 rounded-lg border border-border/30 bg-white px-3 text-xs outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all">
                      <option value="H">Hadamard (H)</option>
                      <option value="X">Pauli-X (X)</option>
                      <option value="Y">Pauli-Y (Y)</option>
                      <option value="Z">Pauli-Z (Z)</option>
                      <option value="CNOT">CNOT</option>
                    </select>
                    {qubits.length > 1 && (
                      <button onClick={() => removeQubit(q.id)}
                        className="w-7 h-7 rounded-lg hover:bg-red-50 text-muted-foreground/40 hover:text-red-500 flex items-center justify-center transition-colors">
                        ✕
                      </button>
                    )}
                  </div>
                ))}
              </div>
              <button onClick={simulate} disabled={running}
                className="mt-6 w-full py-3 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2">
                {running ? (
                  <><div className="w-4 h-4 rounded-full border-2 border-white/30 border-t-white animate-spin" /> {t('Running...')}</>
                ) : (
                  <>{t('Run Circuit')}</>
                )}
              </button>
            </div>
          </div>

          {/* Results */}
          <div className="w-full lg:w-72">
            <div className="qc-fade-up rounded-2xl bg-white/70 backdrop-blur-sm border border-border/10 p-6">
              <h3 className="text-sm font-bold tracking-tight mb-4">{t('Result')}</h3>
              {result ? (
                <div className="text-center py-8">
                  <div className="text-4xl font-bold text-foreground mb-3">{result}</div>
                  <div className="text-xs text-muted-foreground/50">
                    {t('Measured at')} {new Date().toLocaleTimeString()}
                  </div>
                </div>
              ) : running ? (
                <div className="text-center py-8">
                  <div className="w-8 h-8 rounded-full border-2 border-amber-500/30 border-t-amber-500 animate-spin mx-auto mb-3" />
                  <p className="text-xs text-muted-foreground/60">{t('Executing circuit...')}</p>
                </div>
              ) : (
                <div className="text-center py-8">
                  <div className="text-3xl mb-3 opacity-20">⚛</div>
                  <p className="text-xs text-muted-foreground/50">{t('Run a circuit to see results')}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
