import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import Sites from './components/Sites';
import Analytics from './components/Analytics';
import Realtime from './components/Realtime';
import Visits from './components/Visits';
import ChangePassword from './components/ChangePassword';
import About from './components/About';
import { setupAxiosInterceptors } from './utils/axiosConfig';

const theme = createTheme({
  palette: {
    primary: { main: '#1976d2' },
    secondary: { main: '#dc004e' },
  },
});

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const handleUnauthorized = () => setUser(null);

  useEffect(() => {
    setupAxiosInterceptors(handleUnauthorized);
    const token = localStorage.getItem('token');
    const userData = localStorage.getItem('user');
    if (token && userData) {
      setUser(JSON.parse(userData));
    }
    setLoading(false);
  }, []);

  const login = (token, userData) => {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    if (userData.is_default_password) {
      localStorage.setItem('needs_password_change', 'true');
    } else {
      localStorage.removeItem('needs_password_change');
    }
    setUser(userData);
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
  };

  if (loading) return <div>Loading...</div>;

  const guard = (Component) => user
    ? <Component user={user} onLogout={logout} />
    : <Navigate to="/admin/login" />;

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Routes>
          <Route path="/admin/login" element={user ? <Navigate to="/admin" /> : <Login onLogin={login} />} />
          <Route path="/admin" element={user ? <Navigate to="/admin/dashboard" /> : <Navigate to="/admin/login" />} />
          <Route path="/admin/dashboard" element={guard(Dashboard)} />
          <Route path="/admin/realtime" element={guard(Realtime)} />
          <Route path="/admin/sites" element={guard(Sites)} />
          <Route path="/admin/analytics" element={guard(Analytics)} />
          <Route path="/admin/visits" element={guard(Visits)} />
          <Route path="/admin/change-password" element={guard(ChangePassword)} />
          <Route path="/admin/about" element={guard(About)} />
          <Route path="/" element={<Navigate to="/admin" />} />
          <Route path="*" element={<Navigate to="/admin" />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;
