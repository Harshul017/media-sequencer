import { useState } from 'react';
import { triggerSync } from '../api';

export default function SyncForm({ syncState }) {
  const [type, setType] = useState('image');
  const [url, setUrl] = useState('');
  const [duration, setDuration] = useState('10');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await triggerSync({
        type,
        url: url.trim(),
        duration_seconds: Number(duration),
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="panel-form" onSubmit={handleSubmit}>
      <h2 className="panel-title">Trigger sync</h2>
      <p className="panel-hint">
        Every window switches to this item at the same time, then resumes
        its own sequence once the duration ends.
      </p>
      {error && <div className="banner banner--error">{error}</div>}
      {syncState?.active && (
        <div className="banner banner--active">Sync is currently active.</div>
      )}

      <label className="field">
        <span>Type</span>
        <select value={type} onChange={(e) => setType(e.target.value)}>
          <option value="image">Image</option>
          <option value="video">Video</option>
        </select>
      </label>

      <label className="field">
        <span>Media URL</span>
        <input
          type="text"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://…"
          required
        />
      </label>

      <label className="field">
        <span>Duration (seconds)</span>
        <input
          type="number"
          min="1"
          value={duration}
          onChange={(e) => setDuration(e.target.value)}
          required
        />
      </label>

      <button className="btn-accent" type="submit" disabled={submitting}>
        {submitting ? 'Syncing…' : 'Sync all windows'}
      </button>
    </form>
  );
}
