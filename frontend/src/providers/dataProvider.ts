import { DataProvider } from '@refinedev/core';

const API_BASE_URL = '/api/v1';

export const erpDataProvider: DataProvider = {
  getList: async ({ resource, pagination, filters, sorters }) => {
    try {
      const { current = 1, pageSize = 20 } = pagination ?? {};
      const response = await fetch(`${API_BASE_URL}/${resource}?page=${current}&per_page=${pageSize}`);
      if (!response.ok) {
        throw new Error(`Failed to fetch ${resource}`);
      }
      const data = await response.json();
      return {
        data: Array.isArray(data) ? data : data.items || [],
        total: data.total || (Array.isArray(data) ? data.length : 0),
      };
    } catch (error) {
      // Return empty array on network failure to maintain UI stability
      return { data: [], total: 0 };
    }
  },

  getOne: async ({ resource, id }) => {
    const response = await fetch(`${API_BASE_URL}/${resource}/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch ${resource} with id ${id}`);
    }
    const data = await response.json();
    return { data };
  },

  create: async ({ resource, variables }) => {
    const response = await fetch(`${API_BASE_URL}/${resource}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(variables),
    });
    if (!response.ok) {
      throw new Error(`Failed to create ${resource}`);
    }
    const data = await response.json();
    return { data };
  },

  update: async ({ resource, id, variables }) => {
    const response = await fetch(`${API_BASE_URL}/${resource}/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(variables),
    });
    if (!response.ok) {
      throw new Error(`Failed to update ${resource} with id ${id}`);
    }
    const data = await response.json();
    return { data };
  },

  deleteOne: async ({ resource, id }) => {
    const response = await fetch(`${API_BASE_URL}/${resource}/${id}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      throw new Error(`Failed to delete ${resource} with id ${id}`);
    }
    return { data: { id } as any };
  },

  getApiUrl: () => API_BASE_URL,

  custom: async ({ url, method, payload }) => {
    const response = await fetch(url, {
      method: method || 'GET',
      headers: { 'Content-Type': 'application/json' },
      body: payload ? JSON.stringify(payload) : undefined,
    });
    if (!response.ok) {
      throw new Error(`Custom request failed for ${url}`);
    }
    const data = await response.json();
    return { data };
  },
};
