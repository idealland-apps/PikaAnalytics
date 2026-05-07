import React, { useState } from 'react';
import { 
  Container, 
  Paper, 
  TextField, 
  Button, 
  Typography, 
  Box,
  Alert
} from '@mui/material';
import Header from './Header';
import axios from 'axios';
import config from '../config';

const API_BASE_URL = config.API_BASE_URL;

function ChangePassword({ user, onLogout }) {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    // Validate passwords match
    if (newPassword !== confirmPassword) {
      setError('New passwords do not match');
      setLoading(false);
      return;
    }

    // Validate password length
    if (newPassword.length < 6) {
      setError('New password must be at least 6 characters long');
      setLoading(false);
      return;
    }

    try {
      const response = await axios.post(`${API_BASE_URL}/change-password`, {
        current_password: currentPassword,
        new_password: newPassword
      });

      setSuccess('Password changed successfully!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      
      // Update the default password flag based on backend response
      if (response.data.is_default_password) {
        localStorage.setItem('needs_password_change', 'true');
        // Add warning if they set it back to default
        setSuccess('Password changed successfully! However, you are still using the default password. Please consider using a more secure password.');
      } else {
        localStorage.removeItem('needs_password_change');
      }
      
      // Dispatch custom event to notify Header component
      window.dispatchEvent(new CustomEvent('passwordChanged', { 
        detail: { isDefaultPassword: response.data.is_default_password } 
      }));
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to change password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Header 
        user={user} 
        onLogout={onLogout} 
        title="Change Password"
        showBackButton={true}
        backButtonText="Back"
        backDestination="/admin"
      />

      <Container maxWidth="sm">
        <Box
          sx={{
            marginTop: 4,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
          }}
        >
          <Paper elevation={3} sx={{ padding: 4, width: '100%' }}>
            <Typography component="h1" variant="h4" align="center" gutterBottom>
              Change Password
            </Typography>
            
            {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
            {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
            
            <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
              <TextField
                margin="normal"
                required
                fullWidth
                name="currentPassword"
                label="Current Password"
                type="password"
                id="currentPassword"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
              />
              <TextField
                margin="normal"
                required
                fullWidth
                name="newPassword"
                label="New Password"
                type="password"
                id="newPassword"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                helperText="Password must be at least 6 characters long"
              />
              <TextField
                margin="normal"
                required
                fullWidth
                name="confirmPassword"
                label="Confirm New Password"
                type="password"
                id="confirmPassword"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                sx={{ mt: 3, mb: 2 }}
                disabled={loading}
              >
                {loading ? 'Changing Password...' : 'Change Password'}
              </Button>
            </Box>
          </Paper>
        </Box>
      </Container>
    </>
  );
}

export default ChangePassword;