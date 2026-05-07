import React, { useState, useEffect, useCallback } from 'react';
import {
  Container, Typography, Box, Paper, MenuItem, Select, FormControl, InputLabel,
  CircularProgress, Alert, Grid,
} from '@mui/material';
import axios from 'axios';
import config from '../config';
import Header from './Header';
import MonthSelector from './MonthSelector';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement,
  Title, Tooltip as ChartTooltip, Legend, Filler,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, ChartTooltip, Legend, Filler);

const API_BASE_URL = config.API_BASE_URL;

function StatCard({ label, value }) {
  return (
    <Paper sx={{ p: 3, textAlign: 'center' }}>
      <Typography variant="subtitle2" color="text.secondary">{label}</Typography>
      <Typography variant="h4">{value ?? '—'}</Typography>
    </Paper>
  );
}

function formatDuration(seconds) {
  if (!seconds || seconds <= 0) return '0s';
  const s = Math.round(seconds);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  const rem = s % 60;
  return rem ? `${m}m ${rem}s` : `${m}m`;
}

function formatPercent(rate) {
  if (rate == null) return '—';
  return `${(rate * 100).toFixed(1)}%`;
}

function Dashboard({ user, onLogout }) {
  const [sites, setSites] = useState([]);
  const [siteKey, setSiteKey] = useState('');
  const [month, setMonth] = useState('');
  const [overview, setOverview] = useState(null);
  const [visits, setVisits] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const params = useCallback(() => {
    const p = {};
    if (month) p.month = month;
    if (siteKey) p.site_key = siteKey;
    return { params: p };
  }, [siteKey, month]);

  useEffect(() => {
    axios.get(`${API_BASE_URL}/sites`)
      .then((r) => setSites(r.data || []))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!month) return;
    let cancelled = false;
    setLoading(true);
    setError('');
    Promise.all([
      axios.get(`${API_BASE_URL}/analytics/overview`, params()),
      axios.get(`${API_BASE_URL}/analytics/visits`, params()),
    ])
      .then(([o, v]) => {
        if (cancelled) return;
        setOverview(o.data);
        setVisits(v.data.visits || []);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err.response?.data?.error || 'Failed to load dashboard');
      })
      .finally(() => !cancelled && setLoading(false));
    return () => { cancelled = true; };
  }, [params]);

  const chartData = {
    labels: visits.map((v) => v.date),
    datasets: [
      {
        label: 'Page Views',
        data: visits.map((v) => v.views),
        borderColor: '#1976d2',
        backgroundColor: 'rgba(25, 118, 210, 0.18)',
        fill: true,
        tension: 0.3,
      },
      {
        label: 'Unique Visitors',
        data: visits.map((v) => v.uniques),
        borderColor: '#dc004e',
        backgroundColor: 'rgba(220, 0, 78, 0.12)',
        fill: false,
        tension: 0.3,
      },
    ],
  };

  return (
    <>
      <Header user={user} onLogout={onLogout} title="PikaAnalytics Dashboard" />
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, flexWrap: 'wrap', gap: 2 }}>
          <Typography variant="h4">Dashboard</Typography>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <FormControl size="small" sx={{ minWidth: 180 }}>
              <InputLabel>Site</InputLabel>
              <Select value={siteKey} label="Site" onChange={(e) => setSiteKey(e.target.value)}>
                <MenuItem value="">All Sites</MenuItem>
                {sites.map((s) => (
                  <MenuItem key={s.id} value={s.site_key}>{s.name}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <MonthSelector value={month} onChange={setMonth} />
          </Box>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        {loading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
            <CircularProgress />
          </Box>
        )}

        {!loading && overview && (
          <>
            <Grid container spacing={3}>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Page Views" value={overview.total_views} /></Grid>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Unique Visitors" value={overview.unique_ips} /></Grid>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Sessions" value={overview.sessions} /></Grid>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Bounce Rate" value={formatPercent(overview.bounce_rate)} /></Grid>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Avg. Visit Time" value={formatDuration(overview.avg_session_duration)} /></Grid>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Unique Pages" value={overview.unique_pages} /></Grid>
              <Grid item xs={12} sm={6} md={3}><StatCard label="Unique Referrers" value={overview.unique_referrers} /></Grid>
            </Grid>

            <Paper sx={{ p: 3, mt: 3 }}>
              <Typography variant="h6" gutterBottom>Visits Trend</Typography>
              {visits.length === 0 ? (
                <Typography color="text.secondary">No traffic in selected month.</Typography>
              ) : (
                <Box sx={{ height: 320 }}>
                  <Line data={chartData} options={{ maintainAspectRatio: false, responsive: true }} />
                </Box>
              )}
            </Paper>
          </>
        )}
      </Container>
    </>
  );
}

export default Dashboard;
