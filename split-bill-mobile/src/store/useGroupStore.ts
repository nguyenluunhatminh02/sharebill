import {create} from 'zustand';
import {Group, CreateGroupRequest} from '../types';
import {groupAPI} from '../api/services';

interface GroupState {
  groups: Group[];
  currentGroup: Group | null;
  isLoading: boolean;
  error: string | null;

  fetchGroups: () => Promise<void>;
  fetchGroup: (id: string) => Promise<void>;
  createGroup: (data: CreateGroupRequest) => Promise<Group>;
  joinGroup: (inviteCode: string) => Promise<Group>;
  deleteGroup: (id: string) => Promise<void>;
  setCurrentGroup: (group: Group | null) => void;
}

export const useGroupStore = create<GroupState>((set, get) => ({
  groups: [],
  currentGroup: null,
  isLoading: false,
  error: null,

  fetchGroups: async () => {
    try {
      set({isLoading: true, error: null});
      const response = await groupAPI.list();
      if (response.success) {
        set({groups: response.data || []});
      }
    } catch (error: any) {
      set({error: error.message});
    } finally {
      set({isLoading: false});
    }
  },

  fetchGroup: async (id: string) => {
    try {
      set({isLoading: true, error: null});
      const response = await groupAPI.getById(id);
      if (response.success) {
        set({currentGroup: response.data});
      }
    } catch (error: any) {
      set({error: error.message});
    } finally {
      set({isLoading: false});
    }
  },

  createGroup: async (data: CreateGroupRequest) => {
    try {
      set({isLoading: true, error: null});
      const response = await groupAPI.create(data);
      if (response.success) {
        const groups = get().groups;
        set({groups: [response.data, ...groups]});
        return response.data;
      }
      throw new Error('Failed to create group');
    } catch (error: any) {
      set({error: error.message});
      throw error;
    } finally {
      set({isLoading: false});
    }
  },

  joinGroup: async (inviteCode: string) => {
    try {
      set({isLoading: true, error: null});
      const response = await groupAPI.join(inviteCode);
      if (response.success) {
        await get().fetchGroups();
        return response.data;
      }
      throw new Error('Failed to join group');
    } catch (error: any) {
      set({error: error.message});
      throw error;
    } finally {
      set({isLoading: false});
    }
  },

  deleteGroup: async (id: string) => {
    try {
      await groupAPI.delete(id);
      const groups = get().groups.filter(g => g.id !== id);
      set({groups});
    } catch (error: any) {
      set({error: error.message});
      throw error;
    }
  },

  setCurrentGroup: (group: Group | null) => {
    set({currentGroup: group});
  },
}));
