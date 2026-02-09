import {create} from 'zustand';
import {Bill, Balance, Settlement, CreateBillRequest} from '../types';
import {billAPI} from '../api/services';

interface BillState {
  bills: Bill[];
  currentBill: Bill | null;
  balances: Balance[];
  settlements: Settlement[];
  isLoading: boolean;
  error: string | null;

  fetchBills: (groupId: string) => Promise<void>;
  fetchBill: (id: string) => Promise<void>;
  createBill: (groupId: string, data: CreateBillRequest) => Promise<Bill>;
  deleteBill: (id: string) => Promise<void>;
  fetchBalances: (groupId: string) => Promise<void>;
  fetchSettlements: (groupId: string) => Promise<void>;
  setCurrentBill: (bill: Bill | null) => void;
  clearBills: () => void;
}

export const useBillStore = create<BillState>((set, get) => ({
  bills: [],
  currentBill: null,
  balances: [],
  settlements: [],
  isLoading: false,
  error: null,

  fetchBills: async (groupId: string) => {
    try {
      set({isLoading: true, error: null});
      const response = await billAPI.listByGroup(groupId);
      if (response.success) {
        set({bills: response.data || []});
      }
    } catch (error: any) {
      set({error: error.message});
    } finally {
      set({isLoading: false});
    }
  },

  fetchBill: async (id: string) => {
    try {
      set({isLoading: true});
      const response = await billAPI.getById(id);
      if (response.success) {
        set({currentBill: response.data});
      }
    } catch (error: any) {
      set({error: error.message});
    } finally {
      set({isLoading: false});
    }
  },

  createBill: async (groupId: string, data: CreateBillRequest) => {
    try {
      set({isLoading: true, error: null});
      const response = await billAPI.create(groupId, data);
      if (response.success) {
        const bills = get().bills;
        set({bills: [response.data, ...bills]});
        return response.data;
      }
      throw new Error('Failed to create bill');
    } catch (error: any) {
      set({error: error.message});
      throw error;
    } finally {
      set({isLoading: false});
    }
  },

  deleteBill: async (id: string) => {
    try {
      await billAPI.delete(id);
      const bills = get().bills.filter((b: Bill) => b.id !== id);
      set({bills});
    } catch (error: any) {
      set({error: error.message});
      throw error;
    }
  },

  fetchBalances: async (groupId: string) => {
    try {
      const response = await billAPI.getBalances(groupId);
      if (response.success) {
        set({balances: response.data || []});
      }
    } catch (error: any) {
      set({error: error.message});
    }
  },

  fetchSettlements: async (groupId: string) => {
    try {
      const response = await billAPI.getSettlements(groupId);
      if (response.success) {
        set({settlements: response.data || []});
      }
    } catch (error: any) {
      set({error: error.message});
    }
  },

  setCurrentBill: (bill: Bill | null) => set({currentBill: bill}),

  clearBills: () => set({bills: [], balances: [], settlements: []}),
}));
