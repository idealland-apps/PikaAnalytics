import React, { useState, useEffect } from 'react';
import { Container, Paper, Typography, Box, CircularProgress, Alert } from '@mui/material';
import { Info } from '@mui/icons-material';
import Header from './Header';
import axios from 'axios';
import config from '../config';

function About({ user, onLogout }) {
  const [version, setVersion] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    axios.get(`${config.API_BASE_URL}/version`)
      .then((r) => setVersion(r.data.version))
      .catch(() => setError('Failed to load version information'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <Header user={user} onLogout={onLogout} title="About PikaAnalytics" showBackButton={true} />
      <Container maxWidth="md" sx={{ mt: 4 }}>
        <Paper elevation={3} sx={{ p: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
            <Info sx={{ fontSize: 40, mr: 2, color: 'primary.main' }} />
            <Typography variant="h5" component="h1">About PikaAnalytics</Typography>
          </Box>

          {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

          <Typography variant="body1" sx={{ mb: 3 }}>
            PikaAnalytics is a self-hosted website analytics tool. Configure sites, embed the lightweight tracking
            script, and review traffic, devices, referrers, and locations from this admin console.
          </Typography>

          <Box sx={{ mb: 3 }}>
            <Typography variant="h6" gutterBottom>Version</Typography>
            {loading ? (
              <Box sx={{ display: 'flex', alignItems: 'center' }}>
                <CircularProgress size={20} sx={{ mr: 2 }} />
                <Typography>Loading...</Typography>
              </Box>
            ) : (
              <Typography><strong>Version:</strong> {version || 'Unknown'}</Typography>
            )}
          </Box>

          <Box>
            <Typography variant="h6" gutterBottom>Default Credentials</Typography>
            <Typography>Username: <code>admin</code> · Password: <code>admin123</code></Typography>
            <Typography variant="caption" color="text.secondary">Change the password after first login.</Typography>
          </Box>
        </Paper>
      </Container>
    </div>
  );
}

export default About;
