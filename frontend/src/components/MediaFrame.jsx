export default function MediaFrame({ type, url }) {
  if (type === 'blank' || !url) {
    return <div className="media-frame media-frame--blank" aria-label="Blank" />;
  }
  if (type === 'video') {
    return (
      <video
        className="media-frame"
        src={url}
        autoPlay
        muted
        loop
        playsInline
      />
    );
  }
  return <img className="media-frame" src={url} alt="" />;
}
