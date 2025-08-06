import {BrowserRouter as Router, Routes, Route} from "react-router-dom";
import Home from "./pages/Home/Home";
import Dashboard from "./pages/Dashboard/Dashboard";
import {AuthProvider} from "./contexts/AuthContext";

function App() {
    return (
        <AuthProvider>
            <Router>
                <Routes>
                    <Route path="/" element={<Home/>}/>
                    <Route path="/dashboard" element={<Dashboard/>}/>
                </Routes>
            </Router>
        </AuthProvider>
    );
}

export default App;
