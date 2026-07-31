import { Routes, Route } from "react-router-dom";
import { HomePage } from "./components/HomePage";
import { WordInfoPage } from "./components/WordInfoPage";
import { TranslationProvider } from "./context/TranslatorContext";
import { AIExamplesProvider } from "./context/AIExamplesContext";

function App() {
  return (
    <TranslationProvider>
      <AIExamplesProvider>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/info/:word" element={<WordInfoPage />} />
        </Routes>
      </AIExamplesProvider>
    </TranslationProvider>
  );
}

export default App;
