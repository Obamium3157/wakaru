import { Routes, Route } from "react-router-dom";
import { HomePage } from "./components/HomePage";
import { WordInfoPage } from "./components/WordInfoPage";

function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/info/:word" element={<WordInfoPage />} />
    </Routes>
  );
}

export default App;
