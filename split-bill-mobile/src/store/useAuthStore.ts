import {create} from 'zustand';
import {User} from '../types';
import {authAPI} from '../api/services';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  setToken: (token: string) => void;
  setUser: (user: User) => void;
  verifyToken: () => Promise<void>;
  updateProfile: (data: Partial<User>) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: false,

  setToken: (token: string) => {
    set({token, isAuthenticated: true});
  },

  setUser: (user: User) => {
    set({user});
  },

  verifyToken: async () => {
    try {
      set({isLoading: true});
      const response = await authAPI.verifyToken();
      if (response.success) {
        set({user: response.data, isAuthenticated: true});
      }
    } catch (error) {
      console.error('Token verification failed:', error);
      set({isAuthenticated: false, token: null});
    } finally {
      set({isLoading: false});
    }
  },

  updateProfile: async (data: Partial<User>) => {
    try {
      const response = await authAPI.updateProfile(data);
      if (response.success) {
        set({user: response.data});
      }
    } catch (error) {
      console.error('Profile update failed:', error);
      throw error;
    }
  },

  logout: () => {
    set({user: null, token: null, isAuthenticated: false});
  },
}));
