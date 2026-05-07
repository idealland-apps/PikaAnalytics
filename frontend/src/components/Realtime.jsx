import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Container, Typography, Box, Paper, MenuItem, Select, FormControl, InputLabel,
  CircularProgress, Alert, Grid, List, ListItem, ListItemText, Chip,
} from '@mui/material';
import { FiberManualRecord } from '@mui/icons-material';
import axios from 'axios';
import config from '../config';
import Header from './Header';
import { Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS, CategoryScale, LinearScale, BarElement,
  Title, Tooltip as ChartTooltip, Legend,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, ChartTooltip, Legend);

const API_BASE_URL = config.API_BASE_URL;
const REFRESH_MS = 10000;

function StatCard({ label, value, accent }) {
  return (
    <Paper sx={{ p: 3, textAlign: 'center' }}>
      <Typography variant="subtitle2" color="text.secondary">{label}</Typography>
      <Typography variant="h3" sx={{ color: accent }}>{value ?? '—'}</Typography>
    </Paper>
  );
}

function TopList({ title, items, emptyText }) {
  return (
    <Paper sx={{ p: 2, height: '100%' }}>
      <Typography variant="h6" gutterBottom>{title}</Typography>
      {(!items || items.length === 0) ? (
        <Typography color="text.secondary" variant="body2">{emptyText}</Typography>
      ) : (
        <List dense>
          {items.map((it, i) => (
            <ListItem key={i} secondaryAction={<Chip size="small" label={it.count} />}>
              <ListItemText
                primary={it.label}
                primaryTypographyProps={{ noWrap: true, title: it.label }}
              />
            </ListItem>
          ))}
        </List>
      )}
    </Paper>
  );
}

function Realtime({ user, onLogout }) {
  const [sites, setSites] = useState([]);
  const [siteKey, setSiteKey] = useState('');
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const timerRef = useRef(null);

  const fetchData = useCallback(() => {
    const params = {};
    if (siteKey) params.site_key = siteKey;
    return axios.get(`${API_BASE_URL}/analytics/realtime`, { params })
      .then((r) => { setData(r.data); setError(''); })
      .catch((e) => setError(e.response?.data?.error || 'Failed to load realtime data'));
  }, [siteKey]);

  useEffect(() => {
    axios.get(`${API_BASE_URL}/sites`).then((r) => setSites(r.data || [])).catch(() => {});
  }, []);

  useEffect(() => {
    setLoading(true);
    fetchData().finally(() => setLoading(false));
    timerRef.current = setInterval(fetchData, REFRESH_MS);
    return () => clearInterval(timerRef.current);
  }, [fetchData]);

  const buckets = data?.buckets || [];
  const chartData = {
    labels: buckets.map((b) => b.minute.slice(11, 16)),
    datasets: [{
      label: 'Pageviews / minute',
      data: buckets.map((b) => b.views),
      backgroundColor: 'rgba(76, 175, 80, 0.7)',
    }],
  };

  return (
    <>
      <Header user={user} onLogout={onLogout} title="PikaAnalytics Realtime" />
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, flexWrap: 'wrap', gap: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <FiberManualRecord sx={{ color: '#4caf50', fontSize: 16 }} />
            <Typography variant="h4">Realtime</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ ml: 1 }}>
              Last 5 minutes · auto-refresh {REFRESH_MS / 1000}s
            </Typography>
          </Box>
          <FormControl size="small" sx={{ minWidth: 200 }}>
            <InputLabel>Site</InputLabel>
            <Select value={siteKey} label="Site" onChange={(e) => setSiteKey(e.target.value)}>
              <MenuItem value="">All Sites</MenuItem>
              {sites.map((s) => (
                <MenuItem key={s.id} value={s.site_key}>{s.name}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        {loading && !data && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}><CircularProgress /></Box>
        )}

        {data && (
          <>
            <Grid container spacing={3} sx={{ mb: 3 }}>
              <Grid item xs={12} sm={6}>
                <StatCard label="Active Visitors (last 5 min)" value={data.active_visitors} accent="#4caf50" />
              </Grid>
              <Grid item xs={12} sm={6}>
                <StatCard label="Pageviews (last 5 min)" value={data.pageviews} />
              </Grid>
            </Grid>

            <Paper sx={{ p: 3, mb: 3 }}>
              <Typography variant="h6" gutterBottom>Pageviews — Last 30 minutes</Typography>
              {buckets.length === 0 ? (
                <Typography color="text.secondary">No traffic in the last 30 minutes.</Typography>
              ) : (
                <Box sx={{ height: 240 }}>
                  <Bar data={chartData} options={{
                    maintainAspectRatio: false,
                    responsive: true,
                    plugins: { legend: { display: false } },
                    scales: { y: { beginAtZero: true, ticks: { precision: 0 } } },
                  }} />
                </Box>
              )}
            </Paper>

            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <TopList title="Active Pages" items={data.top_pages} emptyText="No active pages." />
              </Grid>
              <Grid item xs={12} md={6}>
                <TopList title="Top Referrers" items={data.top_referrers} emptyText="No referrers." />
              </Grid>
            </Grid>
          </>
        )}
      </Container>
    </>
  );
}

export default Realtime;
