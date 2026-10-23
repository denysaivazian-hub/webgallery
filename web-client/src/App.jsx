import { useState } from "react";
import ImagePage from "./image-page";

const images = Array.from({ length: 5 }, (_, i) => `/photos/image${i + 1}.jpg`);

export default function App() {
  const [selectedIndex, setSelectedIndex] = useState(null);

  if (selectedIndex !== null) {
    return <ImagePage images={images} initialIndex={selectedIndex} onClose={() => setSelectedIndex(null)} />;
  }

  return (
    <main className="gallery">
      <header className="gallery-header">
        <div>
          <p className="eyebrow">Коллекция фотографий</p>
          <h1>Галерея</h1>
          <p className="subtitle">Моменты, к которым хочется вернуться.</p>
        </div>
        <span className="photo-count">{images.length} фото</span>
      </header>
      <div className="photo-grid">
        {images.map((src, index) => (
          <button className="photo-card" key={src} onClick={() => setSelectedIndex(index)} aria-label={`Открыть фотографию ${index + 1}`}>
            <img src={src} alt={`Фотография ${index + 1}`} loading="lazy" />
            <span className="photo-caption">Фото {String(index + 1).padStart(2, "0")} <span aria-hidden="true">↗</span></span>
          </button>
        ))}
      </div>
    </main>
  );
}
