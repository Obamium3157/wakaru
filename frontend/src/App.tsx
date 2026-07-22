import { Routes, Route, BrowserRouter } from "react-router-dom";
import { WordInfoPage } from "./components/WordInfoPage";
import { HomePage } from "./components/HomePage";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/info/:word" element={<WordInfoPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
