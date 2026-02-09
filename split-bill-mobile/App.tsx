import React, {useEffect, useState} from 'react';
import {
  StatusBar,
  StyleSheet,
  View,
  Text,
  ActivityIndicator,
} from 'react-native';
import {NavigationContainer} from '@react-navigation/native';
import {SafeAreaProvider} from 'react-native-safe-area-context';
import AppNavigator from './src/navigation/AppNavigator';
import {useAuthStore} from './src/store/useAuthStore';
import {colors, fontSize, spacing} from './src/theme';

const App: React.FC = () => {
  const [isReady, setIsReady] = useState(false);
  const {token} = useAuthStore();

  useEffect(() => {
    const init = async () => {
      try {
        // Initialize app - load persisted auth state, etc.
        // Zustand with persist middleware handles this automatically
        // Add any additional initialization logic here
        await new Promise(resolve => setTimeout(resolve, 500)); // Splash delay
      } catch (error) {
        console.error('App initialization error:', error);
      } finally {
        setIsReady(true);
      }
    };

    init();
  }, []);

  if (!isReady) {
    return (
      <View style={styles.splashContainer}>
        <StatusBar
          backgroundColor={colors.primary}
          barStyle="light-content"
        />
        <Text style={styles.splashTitle}>💰 SplitBill</Text>
        <Text style={styles.splashSubtitle}>Split smarter, not harder</Text>
        <ActivityIndicator
          size="large"
          color="#FFFFFF"
          style={styles.splashLoader}
        />
      </View>
    );
  }

  return (
    <SafeAreaProvider>
      <NavigationContainer>
        <StatusBar
          backgroundColor={
            token ? colors.background : colors.primary
          }
          barStyle={token ? 'dark-content' : 'light-content'}
        />
        <AppNavigator />
      </NavigationContainer>
    </SafeAreaProvider>
  );
};

const styles = StyleSheet.create({
  splashContainer: {
    flex: 1,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
  },
  splashTitle: {
    fontSize: 36,
    fontWeight: '800',
    color: '#FFFFFF',
  },
  splashSubtitle: {
    fontSize: fontSize.md,
    color: 'rgba(255,255,255,0.8)',
    marginTop: spacing.sm,
  },
  splashLoader: {
    marginTop: spacing.xl,
  },
});

export default App;
