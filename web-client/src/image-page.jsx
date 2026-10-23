import { useState } from "react";

export default function ImagePage({ images, initialIndex, onClose }) {
  const [index, setIndex] = useState(initialIndex);
  const previous = () => setIndex(current => (current - 1 + images.length) % images.length);
  const next = () => setIndex(current => (current + 1) % images.length);

  return (
    <main className="viewer">
      <header className="viewer-header">
        <button className="back-button" onClick={onClose}>← В галерею</button>
        <span className="viewer-count">{index + 1} / {images.length}</span>
      </header>
      <div className="viewer-stage">
        <button className="arrow arrow-previous" onClick={previous} aria-label="Предыдущее фото">←</button>
        <img className="viewer-image" src={images[index]} alt={`Фотография ${index + 1}`} />
        <button className="arrow arrow-next" onClick={next} aria-label="Следующее фото">→</button>
      </div>
      <div className="thumbnails" aria-label="Выбрать фотографию">
        {images.map((src, i) => (
          <button className={`thumbnail ${i === index ? "is-active" : ""}`} key={src} onClick={() => setIndex(i)} aria-label={`Фотография ${i + 1}`} aria-pressed={i === index}>
            <img src={src} alt="" />
          </button>
        ))}
      </div>
    </main>
  );
}
