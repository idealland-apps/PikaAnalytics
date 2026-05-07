import React, { useState, useEffect, useCallback } from 'react';
import {
  Container, Typography, Box, Paper, Grid, Alert, CircularProgress,
  MenuItem, Select, FormControl, InputLabel, LinearProgress,
} from '@mui/material';
import axios from 'axios';
import config from '../config';
import Header from './Header';
import MonthSelector from './MonthSelector';
import { Bar, Doughnut } from 'react-chartjs-2';
import {
  Chart as ChartJS, CategoryScale, LinearScale, BarElement, ArcElement,
  Tooltip as ChartTooltip, Legend,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, ChartTooltip, Legend);

const API_BASE_URL = config.API_BASE_URL;
const COLORS = ['#1976d2', '#dc004e', '#2e7d32', '#ed6c02', '#9c27b0', '#0288d1', '#d32f2f', '#5d4037'];

function TopList({ title, items }) {
  const total = (items || []).reduce((acc, i) => acc + i.count, 0) || 1;
  return (
    <Paper sx={{ p: 3, height: '100%' }}>
      <Typography variant="h6" gutterBottom>{title}</Typography>
      {(!items || items.length === 0) && <Typography color="text.secondary">No data</Typography>}
      {items && items.map((item) => (
        <Box key={item.label} sx={{ mb: 1.5 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
            <Typography variant="body2" sx={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '70%' }}>{item.label}</Typography>
            <Typography variant="body2" color="text.secondary">{item.count}</Typography>
          </Box>
          <LinearProgress variant="determinate" value={(item.count / total) * 100} sx={{ height: 6, borderRadius: 3 }} />
        </Box>
      ))}
    </Paper>
  );
}

function Analytics({ user, onLogout }) {
  const [sites, setSites] = useState([]);
  const [siteKey, setSiteKey] = useState('');
  const [month, setMonth] = useState('');
  const [data, setData] = useState({ pages: null, referrers: null, devices: null, locations: null });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const params = useCallback(() => {
    const p = {};
    if (month) p.month = month;
    if (siteKey) p.site_key = siteKey;
    return { params: p };
  }, [siteKey, month]);

  useEffect(() => {
    axios.get(`${API_BASE_URL}/sites`).then((r) => setSites(r.data || [])).catch(() => {});
  }, []);

  useEffect(() => {
    if (!month) return;
    let cancelled = false;
    setLoading(true);
    setError('');
    Promise.all([
      axios.get(`${API_BASE_URL}/analytics/pages`, params()),
      axios.get(`${API_BASE_URL}/analytics/referrers`, params()),
      axios.get(`${API_BASE_URL}/analytics/devices`, params()),
      axios.get(`${API_BASE_URL}/analytics/locations`, params()),
    ])
      .then(([p, r, d, l]) => {
        if (cancelled) return;
        setData({
          pages: p.data.pages || [],
          referrers: r.data.referrers || [],
          devices: d.data,
          locations: l.data.locations || [],
        });
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err.response?.data?.error || 'Failed to load analytics');
      })
      .finally(() => !cancelled && setLoading(false));
    return () => { cancelled = true; };
  }, [params]);

  const browsersChart = data.devices && {
    labels: (data.devices.browsers || []).slice(0, 8).map((b) => b.label),
    datasets: [{
      data: (data.devices.browsers || []).slice(0, 8).map((b) => b.count),
      backgroundColor: COLORS,
    }],
  };
  const osChart = data.devices && {
    labels: (data.devices.os || []).slice(0, 8).map((b) => b.label),
    datasets: [{
      data: (data.devices.os || []).slice(0, 8).map((b) => b.count),
      backgroundColor: COLORS,
    }],
  };
  const countriesChart = data.locations && {
    labels: data.locations.slice(0, 10).map((b) => b.label),
    datasets: [{
      label: 'Visits',
      data: data.locations.slice(0, 10).map((b) => b.count),
      backgroundColor: '#1976d2',
    }],
  };

  return (
    <>
      <Header user={user} onLogout={onLogout} title="Analytics" />
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, flexWrap: 'wrap', gap: 2 }}>
          <Typography variant="h4">Analytics</Typography>
          <Box sx={{ display: 'flex', gap: 2 }}>
            <FormControl size="small" sx={{ minWidth: 180 }}>
              <InputLabel>Site</InputLabel>
              <Select value={siteKey} label="Site" onChange={(e) => setSiteKey(e.target.value)}>
                <MenuItem value="">All Sites</MenuItem>
                {sites.map((s) => (<MenuItem key={s.id} value={s.site_key}>{s.name}</MenuItem>))}
              </Select>
            </FormControl>
            <MonthSelector value={month} onChange={setMonth} />
          </Box>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}><CircularProgress /></Box>
        ) : (
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}><TopList title="Top Pages" items={data.pages} /></Grid>
            <Grid item xs={12} md={6}><TopList title="Top Referrers" items={data.referrers} /></Grid>

            <Grid item xs={12} md={6}>
              <Paper sx={{ p: 3, height: '100%' }}>
                <Typography variant="h6" gutterBottom>Browsers</Typography>
                {browsersChart && browsersChart.labels.length > 0 ? (
                  <Box sx={{ height: 280 }}><Doughnut data={browsersChart} options={{ maintainAspectRatio: false, responsive: true }} /></Box>
                ) : <Typography color="text.secondary">No data</Typography>}
              </Paper>
            </Grid>
            <Grid item xs={12} md={6}>
              <Paper sx={{ p: 3, height: '100%' }}>
                <Typography variant="h6" gutterBottom>Operating Systems</Typography>
                {osChart && osChart.labels.length > 0 ? (
                  <Box sx={{ height: 280 }}><Doughnut data={osChart} options={{ maintainAspectRatio: false, responsive: true }} /></Box>
                ) : <Typography color="text.secondary">No data</Typography>}
              </Paper>
            </Grid>

            <Grid item xs={12}>
              <Paper sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>Top Countries</Typography>
                {countriesChart && countriesChart.labels.length > 0 ? (
                  <Box sx={{ height: 320 }}><Bar data={countriesChart} options={{ maintainAspectRatio: false, responsive: true, plugins: { legend: { display: false } } }} /></Box>
                ) : <Typography color="text.secondary">No data</Typography>}
              </Paper>
            </Grid>
          </Grid>
        )}
      </Container>
    </>
  );
}

export default Analytics;
