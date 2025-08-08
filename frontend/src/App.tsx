import {BrowserRouter as Router, Routes, Route} from "react-router-dom";
import Home from "./pages/Home/Home";
import Dashboard from "./pages/Dashboard/Dashboard";
import {AuthProvider} from "./contexts/AuthContext";
import AccountPage from "./pages/Account/AccountPage";

function App() {
    return (
        <AuthProvider>
            <Router>
                <Routes>
                    <Route path="/" element={<Home/>}/>
                    <Route path="/dashboard" element={<Dashboard/>}/>
                    <Route path="/account" element={<AccountPage/>}/>
                </Routes>
            </Router>
        </AuthProvider>
    );
}

export default App;
