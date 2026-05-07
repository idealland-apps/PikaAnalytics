import React, { useEffect, useState } from 'react';
import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';
import axios from 'axios';
import config from '../config';

const API_BASE_URL = config.API_BASE_URL;

function formatMonthLabel(ym) {
  if (!ym) return '';
  const [y, m] = ym.split('-').map(Number);
  if (!y || !m) return ym;
  const d = new Date(Date.UTC(y, m - 1, 1));
  return d.toLocaleString(undefined, { year: 'numeric', month: 'long' });
}

export default function MonthSelector({ value, onChange, minWidth = 160 }) {
  const [months, setMonths] = useState([]);

  useEffect(() => {
    let cancelled = false;
    axios.get(`${API_BASE_URL}/analytics/months`)
      .then((r) => {
        if (cancelled) return;
        const list = r.data?.months || [];
        setMonths(list);
        if (!value && list.length > 0) {
          onChange(r.data.current || list[0]);
        }
      })
      .catch(() => {});
    return () => { cancelled = true; };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <FormControl size="small" sx={{ minWidth }}>
      <InputLabel>Month</InputLabel>
      <Select value={value || ''} label="Month" onChange={(e) => onChange(e.target.value)}>
        {months.map((m) => (
          <MenuItem key={m} value={m}>{formatMonthLabel(m)}</MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}

export { formatMonthLabel };
