import { useState } from 'react';
import { addMedia } from '../api';

export default function AddMediaForm({ windows, onAdded }) {
  const [windowId, setWindowId] = useState('');
  const [type, setType] = useState('image');
  const [url, setUrl] = useState('');
  const [duration, setDuration] = useState('5');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    if (!windowId) {
      setError('Choose a window.');
      return;
    }
    setSubmitting(true);
    try {
      await addMedia(windowId, {
        type,
        url: type === 'blank' ? '' : url.trim(),
        duration_seconds: Number(duration),
      });
      setUrl('');
      onAdded();
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="panel-form" onSubmit={handleSubmit}>
      <h2 className="panel-title">Add media</h2>
      {error && <div className="banner banner--error">{error}</div>}

      <label className="field">
        <span>Window</span>
        <select value={windowId} onChange={(e) => setWindowId(e.target.value)}>
          <option value="">Select a window</option>
          {windows.map((w) => (
            <option key={w.id} value={w.id}>{w.name}</option>
          ))}
        </select>
      </label>

      <label className="field">
        <span>Type</span>
        <select value={type} onChange={(e) => setType(e.target.value)}>
          <option value="image">Image</option>
          <option value="video">Video</option>
          <option value="blank">Blank</option>
        </select>
      </label>

      {type !== 'blank' && (
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
      )}

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

      <button className="btn-primary" type="submit" disabled={submitting}>
        {submitting ? 'Adding…' : 'Add to playlist'}
      </button>
    </form>
  );
}
