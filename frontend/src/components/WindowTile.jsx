import { useEffect, useState } from 'react';
import MediaFrame from './MediaFrame';
import { computeCurrentItem } from '../playback';

export default function WindowTile({ window, cycleSeconds, getEstimatedNow, syncState }) {
  // Re-render on a short tick so the progress bar and item switches
  // happen smoothly, without needing any network call for normal playback.
  const [, forceTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => forceTick((n) => n + 1), 250);
    return () => clearInterval(id);
  }, []);

  const isSynced = syncState?.active;

  // Subtracting total_paused_seconds is what makes normal playback truly
  // pause during a sync rather than silently skipping ahead — see the
  // comment on Manager.Trigger in the backend for the full explanation.
  const pausedMs = (syncState?.total_paused_seconds || 0) * 1000;
  const adjustedNow = getEstimatedNow() - pausedMs;

  const current = isSynced
    ? null
    : computeCurrentItem(window.items, cycleSeconds, adjustedNow);

  const displayType = isSynced ? syncState.type : current?.item.type;
  const displayUrl = isSynced ? syncState.url : current?.item.url;

  const progress = !isSynced && current ? current.msIntoItem / current.itemDurationMs : null;

  return (
    <div className={`tile ${isSynced ? 'tile--synced' : ''}`}>
      <div className="tile__header">
        <span className="tile__name">{window.name}</span>
        {isSynced ? (
          <span className="tile__badge tile__badge--synced">
            <span className="dot" /> Synced
          </span>
        ) : current ? (
          <span className="tile__badge">
            Item {current.index + 1} of {window.items.length}
          </span>
        ) : null}
      </div>

      <div className="tile__stage">
        {displayType ? (
          <MediaFrame type={displayType} url={displayUrl} />
        ) : (
          <div className="media-frame media-frame--empty">No media configured</div>
        )}
      </div>

      <div className="tile__progress">
        <div
          className="tile__progress-fill"
          style={{ width: progress != null ? `${progress * 100}%` : isSynced ? '100%' : '0%' }}
        />
      </div>
    </div>
  );
}
