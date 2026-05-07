import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Box,
  TextField,
  Button,
  Paper,
  Grid,
  CircularProgress,
  Alert,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Tooltip,
  Snackbar,
} from '@mui/material';
import { ContentCopy, Delete, Edit, Code } from '@mui/icons-material';
import axios from 'axios';
import config from '../config';
import Header from './Header';

const API_BASE_URL = config.API_BASE_URL;

function buildEmbedSnippet(siteKey) {
  const base = window.location.protocol + '//' + window.location.host;
  return `<script async src="${base}/track.js?site=${encodeURIComponent(siteKey)}"></script>`;
}

function Sites({ user, onLogout }) {
  const [sites, setSites] = useState([]);
  const [newSite, setNewSite] = useState({ name: '', site_key: '', domain: '', description: '' });
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState(null);
  const [embedSite, setEmbedSite] = useState(null);
  const [snack, setSnack] = useState('');

  const fetchSites = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await axios.get(`${API_BASE_URL}/sites`);
      setSites(response.data || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load site list');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSites();
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setSubmitting(true);
    setError('');
    try {
      await axios.post(`${API_BASE_URL}/sites`, newSite);
      setNewSite({ name: '', site_key: '', domain: '', description: '' });
      fetchSites();
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create site');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this site? All collected analytics for it will be removed.')) return;
    try {
      await axios.delete(`${API_BASE_URL}/sites/${id}`);
      fetchSites();
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to delete site');
    }
  };

  const handleEditSave = async () => {
    try {
      await axios.put(`${API_BASE_URL}/sites/${editing.id}`, {
        name: editing.name,
        domain: editing.domain,
        description: editing.description,
      });
      setEditing(null);
      fetchSites();
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update site');
    }
  };

  const copyToClipboard = async (text, label) => {
    try {
      await navigator.clipboard.writeText(text);
      setSnack(`${label} copied`);
    } catch (e) {
      setSnack('Copy failed');
    }
  };

  return (
    <>
      <Header user={user} onLogout={onLogout} title="Site Management" />
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" gutterBottom>Sites</Typography>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <Grid container spacing={3}>
          <Grid item xs={12} md={5}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>Add Site</Typography>
              <Box component="form" onSubmit={handleSubmit} noValidate>
                <TextField fullWidth required margin="normal" label="Site Name"
                  value={newSite.name} onChange={(e) => setNewSite({ ...newSite, name: e.target.value })} />
                <TextField fullWidth required margin="normal" label="Site Key"
                  helperText="Unique identifier embedded in tracking script (letters, digits, - and _)"
                  value={newSite.site_key} onChange={(e) => setNewSite({ ...newSite, site_key: e.target.value })} />
                <TextField fullWidth margin="normal" label="Domain (optional)"
                  value={newSite.domain} onChange={(e) => setNewSite({ ...newSite, domain: e.target.value })} />
                <TextField fullWidth margin="normal" label="Description (optional)" multiline rows={2}
                  value={newSite.description} onChange={(e) => setNewSite({ ...newSite, description: e.target.value })} />
                <Button type="submit" variant="contained" disabled={submitting} sx={{ mt: 2 }}>
                  {submitting ? 'Creating…' : 'Create Site'}
                </Button>
              </Box>
            </Paper>
          </Grid>

          <Grid item xs={12} md={7}>
            <Paper sx={{ p: 3, minHeight: 400 }}>
              <Typography variant="h6" gutterBottom>Tracked Sites</Typography>
              {loading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
                  <CircularProgress />
                </Box>
              ) : sites.length === 0 ? (
                <Typography color="text.secondary">No tracked sites yet.</Typography>
              ) : (
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>Name</TableCell>
                      <TableCell>Key</TableCell>
                      <TableCell>Domain</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {sites.map((site) => (
                      <TableRow key={site.id} hover>
                        <TableCell>{site.name}</TableCell>
                        <TableCell><code>{site.site_key}</code></TableCell>
                        <TableCell>{site.domain || '—'}</TableCell>
                        <TableCell align="right">
                          <Tooltip title="Embed code">
                            <IconButton size="small" onClick={() => setEmbedSite(site)}><Code fontSize="small" /></IconButton>
                          </Tooltip>
                          <Tooltip title="Edit">
                            <IconButton size="small" onClick={() => setEditing({ ...site })}><Edit fontSize="small" /></IconButton>
                          </Tooltip>
                          <Tooltip title="Delete">
                            <IconButton size="small" onClick={() => handleDelete(site.id)}><Delete fontSize="small" /></IconButton>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </Paper>
          </Grid>
        </Grid>
      </Container>

      <Dialog open={!!editing} onClose={() => setEditing(null)} fullWidth maxWidth="sm">
        <DialogTitle>Edit Site</DialogTitle>
        <DialogContent>
          {editing && (
            <>
              <TextField fullWidth margin="normal" label="Site Name"
                value={editing.name} onChange={(e) => setEditing({ ...editing, name: e.target.value })} />
              <TextField fullWidth margin="normal" label="Domain"
                value={editing.domain || ''} onChange={(e) => setEditing({ ...editing, domain: e.target.value })} />
              <TextField fullWidth margin="normal" label="Description" multiline rows={2}
                value={editing.description || ''} onChange={(e) => setEditing({ ...editing, description: e.target.value })} />
              <Typography variant="caption" color="text.secondary">Site key cannot be changed.</Typography>
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditing(null)}>Cancel</Button>
          <Button variant="contained" onClick={handleEditSave}>Save</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={!!embedSite} onClose={() => setEmbedSite(null)} fullWidth maxWidth="md">
        <DialogTitle>Embed Tracking Code</DialogTitle>
        <DialogContent>
          {embedSite && (
            <>
              <Typography variant="body2" gutterBottom>
                Add this script to the <code>&lt;head&gt;</code> of pages on <strong>{embedSite.name}</strong>.
              </Typography>
              <Paper variant="outlined" sx={{ p: 2, fontFamily: 'monospace', fontSize: 13, mt: 2, wordBreak: 'break-all' }}>
                {buildEmbedSnippet(embedSite.site_key)}
              </Paper>
              <Button startIcon={<ContentCopy />} sx={{ mt: 2 }}
                onClick={() => copyToClipboard(buildEmbedSnippet(embedSite.site_key), 'Snippet')}>
                Copy Snippet
              </Button>
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEmbedSite(null)}>Close</Button>
        </DialogActions>
      </Dialog>

      <Snackbar open={!!snack} autoHideDuration={2000} onClose={() => setSnack('')} message={snack} />
    </>
  );
}

export default Sites;
