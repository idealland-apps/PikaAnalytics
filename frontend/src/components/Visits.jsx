import React, { useState, useEffect, useCallback } from 'react';
import {
  Container, Typography, Box, Paper, Alert, CircularProgress,
  MenuItem, Select, FormControl, InputLabel, Table, TableHead, TableBody,
  TableRow, TableCell, Chip, TextField,
} from '@mui/material';
import axios from 'axios';
import config from '../config';
import Header from './Header';
import MonthSelector from './MonthSelector';

const API_BASE_URL = config.API_BASE_URL;

function Visits({ user, onLogout }) {
  const [sites, setSites] = useState([]);
  const [siteKey, setSiteKey] = useState('');
  const [month, setMonth] = useState('');
  const [limit, setLimit] = useState(100);
  const [search, setSearch] = useState('');
  const [visits, setVisits] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    axios.get(`${API_BASE_URL}/sites`).then((r) => setSites(r.data || [])).catch(() => {});
  }, []);

  const params = useCallback(() => {
    const p = { limit };
    if (month) p.month = month;
    if (siteKey) p.site_key = siteKey;
    return { params: p };
  }, [siteKey, month, limit]);

  useEffect(() => {
    if (!month) return;
    let cancelled = false;
    setLoading(true);
    setError('');
    axios.get(`${API_BASE_URL}/analytics/recent`, params())
      .then((r) => { if (!cancelled) setVisits(r.data.visits || []); })
      .catch((err) => { if (!cancelled) setError(err.response?.data?.error || 'Failed to load visits'); })
      .finally(() => !cancelled && setLoading(false));
    return () => { cancelled = true; };
  }, [params]);

  const filtered = search
    ? visits.filter((v) => {
        const q = search.toLowerCase();
        return (v.path || '').toLowerCase().includes(q) ||
               (v.referrer || '').toLowerCase().includes(q) ||
               (v.country || '').toLowerCase().includes(q) ||
               (v.ip || '').toLowerCase().includes(q);
      })
    : visits;

  return (
    <>
      <Header user={user} onLogout={onLogout} title="Access Logs" />
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, flexWrap: 'wrap', gap: 2 }}>
          <Typography variant="h4">Access Logs</Typography>
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            <TextField size="small" label="Search" value={search} onChange={(e) => setSearch(e.target.value)} sx={{ minWidth: 200 }} />
            <FormControl size="small" sx={{ minWidth: 180 }}>
              <InputLabel>Site</InputLabel>
              <Select value={siteKey} label="Site" onChange={(e) => setSiteKey(e.target.value)}>
                <MenuItem value="">All Sites</MenuItem>
                {sites.map((s) => (<MenuItem key={s.id} value={s.site_key}>{s.name}</MenuItem>))}
              </Select>
            </FormControl>
            <MonthSelector value={month} onChange={setMonth} />
            <FormControl size="small" sx={{ minWidth: 100 }}>
              <InputLabel>Limit</InputLabel>
              <Select value={limit} label="Limit" onChange={(e) => setLimit(e.target.value)}>
                <MenuItem value={50}>50</MenuItem>
                <MenuItem value={100}>100</MenuItem>
                <MenuItem value={250}>250</MenuItem>
                <MenuItem value={500}>500</MenuItem>
              </Select>
            </FormControl>
          </Box>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <Paper>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CircularProgress /></Box>
          ) : (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Time</TableCell>
                  <TableCell>Site</TableCell>
                  <TableCell>Path</TableCell>
                  <TableCell>Referrer</TableCell>
                  <TableCell>Location</TableCell>
                  <TableCell>Browser / OS</TableCell>
                  <TableCell>IP</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filtered.length === 0 && (
                  <TableRow><TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                    <Typography color="text.secondary">No visits found.</Typography>
                  </TableCell></TableRow>
                )}
                {filtered.map((v) => (
                  <TableRow key={v.id} hover>
                    <TableCell>{new Date(v.created_at).toLocaleString()}</TableCell>
                    <TableCell>{v.site_name || v.site_key}</TableCell>
                    <TableCell sx={{ maxWidth: 280, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {v.path}
                      {v.is_bot && <Chip label="bot" size="small" color="warning" sx={{ ml: 1 }} />}
                    </TableCell>
                    <TableCell sx={{ maxWidth: 240, overflow: 'hidden', textOverflow: 'ellipsis' }}>{v.referrer || '—'}</TableCell>
                    <TableCell>{[v.city, v.country].filter(Boolean).join(', ') || '—'}</TableCell>
                    <TableCell>{[v.browser, v.os].filter(Boolean).join(' / ') || '—'}</TableCell>
                    <TableCell>{v.ip}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </Paper>
      </Container>
    </>
  );
}

export default Visits;
