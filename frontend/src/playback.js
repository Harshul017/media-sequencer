// Pure logic, no React here — easy to reason about and test on its own.
//
// Given a window's ordered playlist and the current time, figures out
// which item should be on screen right now, purely from math: no server
// round-trip needed for normal playback.

/**
 * @param {Array<{duration_seconds:number}>} items
 * @param {number} cycleSeconds - the 5-hour loop length from the backend
 * @param {number} nowMs - current time estimate, in epoch milliseconds
 * @returns {{ item: object, index: number, msIntoItem: number, itemDurationMs: number } | null}
 */
export function computeCurrentItem(items, cycleSeconds, nowMs) {
  if (!items || items.length === 0) return null;

  const totalItemsMs = items.reduce((sum, i) => sum + i.duration_seconds * 1000, 0);
  if (totalItemsMs <= 0) return null;

  const cycleMs = cycleSeconds * 1000;

  // Position within the current 5-hour cycle.
  const posInCycleMs = nowMs % cycleMs;

  // The playlist repeats continuously within the cycle, so we only need
  // this item's position within a single pass through the playlist.
  const posInPlaylistMs = posInCycleMs % totalItemsMs;

  let cursor = 0;
  for (let index = 0; index < items.length; index++) {
    const durationMs = items[index].duration_seconds * 1000;
    if (posInPlaylistMs < cursor + durationMs) {
      return {
        item: items[index],
        index,
        msIntoItem: posInPlaylistMs - cursor,
        itemDurationMs: durationMs,
      };
    }
    cursor += durationMs;
  }

  // Floating point edge case landing exactly on the boundary — show the
  // last item rather than showing nothing.
  const lastIndex = items.length - 1;
  return {
    item: items[lastIndex],
    index: lastIndex,
    msIntoItem: 0,
    itemDurationMs: items[lastIndex].duration_seconds * 1000,
  };
}
