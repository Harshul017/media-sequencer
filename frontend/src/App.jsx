import { useCallback, useEffect, useRef, useState } from 'react';
import { getWindows, getSyncState } from './api';
import WindowTile from './components/WindowTile';
import AddMediaForm from './components/AddMediaForm';
import SyncForm from './components/SyncForm';

export default function App() {
  const [windows, setWindows] = useState([]);
  const [cycleSeconds, setCycleSeconds] = useState(18000);
  const [syncState, setSyncState] = useState({ active: false });
  const [error, setError] = useState('');

  // Tracks the offset between this browser's clock and the backend's
  // clock, captured whenever we fetch /windows. Using a ref (not state)
  // because WindowTile reads it every 250ms via getEstimatedNow() and we
  // don't want that to trigger App-level re-renders.
  const clockOffsetMs = useRef(0);

  const getEstimatedNow = useCallback(() => Date.now() + clockOffsetMs.current, []);

  const loadWindows = useCallback(async () => {
    try {
      const data = await getWindows();
      const serverNowMs = new Date(data.server_time).getTime();
      clockOffsetMs.current = serverNowMs - Date.now();
      setWindows(data.windows);
      setCycleSeconds(data.cycle_seconds);
      setError('');
    } catch {
      setError('Could not reach the backend. Is it running?');
    }
  }, []);

  // Poll /windows periodically so dynamically added media shows up
  // without a manual page refresh.
  useEffect(() => {
    loadWindows();
    const id = setInterval(loadWindows, 8000);
    return () => clearInterval(id);
  }, [loadWindows]);

  // Poll /sync-state frequently — this is the one thing that needs to
  // feel responsive, since it's a live operator action.
  useEffect(() => {
    const poll = async () => {
      try {
        const state = await getSyncState();
        setSyncState(state);
      } catch {
        // A single missed poll isn't worth surfacing an error for —
        // the next tick a second later will just try again.
      }
    };
    poll();
    const id = setInterval(poll, 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <div className="app">
      <header className="topbar">
        <h1 className="wordmark">Signal Wall</h1>
        <p className="subtitle">Multi-window media sequencer with sync playback</p>
      </header>

      {error && <div className="banner banner--error banner--page">{error}</div>}

      <div className="layout">
        <div className="grid">
          {windows.map((w) => (
            <WindowTile
              key={w.id}
              window={w}
              cycleSeconds={cycleSeconds}
              getEstimatedNow={getEstimatedNow}
              syncState={syncState}
            />
          ))}
        </div>

        <aside className="controls">
          <AddMediaForm windows={windows} onAdded={loadWindows} />
          <SyncForm syncState={syncState} />
        </aside>
      </div>
    </div>
  );
}
