import { Routes, Route } from "react-router-dom";
import { HomePage } from "./components/HomePage";
import { WordInfoPage } from "./components/WordInfoPage";
import { TranslationProvider } from "./context/TranslatorContext";

function App() {
  return (
    <TranslationProvider>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/info/:word" element={<WordInfoPage />} />
      </Routes>
    </TranslationProvider>
  );
}

export default App;
