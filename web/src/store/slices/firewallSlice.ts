import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { apiClient } from '@api/client';
import type { FirewallRule, FirewallAlias, NATRule } from '@typedefs/firewall';

interface FirewallState {
  rules: FirewallRule[];
  aliases: FirewallAlias[];
  natRules: NATRule[];
  loading: boolean;
  error: string | null;
}

const initialState: FirewallState = {
  rules: [],
  aliases: [],
  natRules: [],
  loading: false,
  error: null,
};

export const fetchRules = createAsyncThunk('firewall/fetchRules', async () => {
  const response = await apiClient.get('/firewall/rules');
  return response.data;
});

export const createRule = createAsyncThunk(
  'firewall/createRule',
  async (rule: Partial<FirewallRule>) => {
    const response = await apiClient.post('/firewall/rules', rule);
    return response.data;
  },
);

export const updateRule = createAsyncThunk(
  'firewall/updateRule',
  async ({ id, ...data }: Partial<FirewallRule> & { id: string }) => {
    const response = await apiClient.put(`/firewall/rules/${id}`, data);
    return response.data;
  },
);

export const deleteRule = createAsyncThunk('firewall/deleteRule', async (id: string) => {
  await apiClient.delete(`/firewall/rules/${id}`);
  return id;
});

export const fetchAliases = createAsyncThunk('firewall/fetchAliases', async () => {
  const response = await apiClient.get('/firewall/aliases');
  return response.data;
});

const firewallSlice = createSlice({
  name: 'firewall',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchRules.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchRules.fulfilled, (state, action: PayloadAction<FirewallRule[]>) => {
        state.loading = false;
        state.rules = action.payload;
      })
      .addCase(fetchRules.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch rules';
      })
      .addCase(createRule.fulfilled, (state, action: PayloadAction<FirewallRule>) => {
        state.rules.push(action.payload);
      })
      .addCase(updateRule.fulfilled, (state, action: PayloadAction<FirewallRule>) => {
        const idx = state.rules.findIndex((r) => r.id === action.payload.id);
        if (idx !== -1) state.rules[idx] = action.payload;
      })
      .addCase(deleteRule.fulfilled, (state, action: PayloadAction<string>) => {
        state.rules = state.rules.filter((r) => r.id !== action.payload);
      })
      .addCase(fetchAliases.fulfilled, (state, action: PayloadAction<FirewallAlias[]>) => {
        state.aliases = action.payload;
      });
  },
});

export default firewallSlice.reducer;
