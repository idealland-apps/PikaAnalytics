import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AppBar, Toolbar, Typography, Button, IconButton, Menu, MenuItem,
  Alert, Box, Collapse,
} from '@mui/material';
import {
  ArrowBack, ExitToApp, AccountCircle, Settings, Assessment,
  Close, Web, Info, Dashboard as DashboardIcon, ListAlt, Bolt,
} from '@mui/icons-material';

function Header({
  user,
  onLogout,
  title = 'PikaAnalytics',
  showBackButton = false,
  backButtonText = 'Back',
  backDestination = '/admin',
}) {
  const [anchorEl, setAnchorEl] = useState(null);
  const [showPasswordAlert, setShowPasswordAlert] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    setShowPasswordAlert(localStorage.getItem('needs_password_change') === 'true');
    const handler = (e) => setShowPasswordAlert(e.detail.isDefaultPassword);
    window.addEventListener('passwordChanged', handler);
    return () => window.removeEventListener('passwordChanged', handler);
  }, []);

  const handleMenuOpen = (e) => setAnchorEl(e.currentTarget);
  const handleMenuClose = () => setAnchorEl(null);
  const go = (path) => () => { handleMenuClose(); navigate(path); };

  return (
    <Box>
      <AppBar position="static">
        <Toolbar>
          {showBackButton && (
            <Button color="inherit" onClick={() => navigate(backDestination)} startIcon={<ArrowBack />} sx={{ mr: 2 }}>
              {backButtonText}
            </Button>
          )}

          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>{title}</Typography>

          <Button color="inherit" startIcon={<DashboardIcon />} onClick={() => navigate('/admin/dashboard')}>Dashboard</Button>
          <Button color="inherit" startIcon={<Bolt />} onClick={() => navigate('/admin/realtime')}>Realtime</Button>
          <Button color="inherit" startIcon={<Web />} onClick={() => navigate('/admin/sites')}>Sites</Button>
          <Button color="inherit" startIcon={<Assessment />} onClick={() => navigate('/admin/analytics')}>Analytics</Button>
          <Button color="inherit" startIcon={<ListAlt />} onClick={() => navigate('/admin/visits')}>Visits</Button>

          <Typography variant="body2" sx={{ mx: 2 }}>{user.username}</Typography>

          <IconButton color="inherit" onClick={handleMenuOpen}><AccountCircle /></IconButton>
          <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
            <MenuItem onClick={go('/admin/change-password')}><Settings sx={{ mr: 1 }} />Change Password</MenuItem>
            <MenuItem onClick={go('/admin/about')}><Info sx={{ mr: 1 }} />About</MenuItem>
            <MenuItem onClick={() => { handleMenuClose(); onLogout(); }}><ExitToApp sx={{ mr: 1 }} />Logout</MenuItem>
          </Menu>
        </Toolbar>
      </AppBar>

      <Collapse in={showPasswordAlert}>
        <Alert
          severity="warning"
          action={
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button color="inherit" size="small" onClick={() => { setShowPasswordAlert(false); navigate('/admin/change-password'); }}>
                Change Now
              </Button>
              <IconButton color="inherit" size="small" onClick={() => setShowPasswordAlert(false)}><Close fontSize="inherit" /></IconButton>
            </Box>
          }
          sx={{ borderRadius: 0 }}
        >
          You are using the default password. Please change it for security reasons.
        </Alert>
      </Collapse>
    </Box>
  );
}

export default Header;
