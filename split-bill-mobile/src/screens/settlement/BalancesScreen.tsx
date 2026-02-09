import React, {useEffect, useState, useCallback} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  RefreshControl,
  Linking,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialCommunityIcons';
import {colors, spacing, borderRadius, fontSize} from '../../theme';
import {useBillStore} from '../../store/useBillStore';
import {useAuthStore} from '../../store/useAuthStore';
import {transactionService} from '../../api/services';
import type {Balance, Settlement} from '../../types';

interface BalancesScreenProps {
  navigation: any;
  route: {
    params: {
      groupId: string;
      groupName: string;
    };
  };
}

const BalancesScreen: React.FC<BalancesScreenProps> = ({
  navigation,
  route,
}) => {
  const {groupId, groupName} = route.params;
  const {user} = useAuthStore();
  const {balances, settlements, loading, fetchBalances, fetchSettlements} =
    useBillStore();
  const [activeTab, setActiveTab] = useState<'balances' | 'settlements'>(
    'balances',
  );
  const [refreshing, setRefreshing] = useState(false);
  const [settling, setSettling] = useState<string | null>(null);

  useEffect(() => {
    navigation.setOptions({title: `${groupName} - Balances`});
    loadData();
  }, [groupId]);

  const loadData = async () => {
    await Promise.all([fetchBalances(groupId), fetchSettlements(groupId)]);
  };

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadData();
    setRefreshing(false);
  }, [groupId]);

  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const handleSettle = async (settlement: Settlement) => {
    Alert.alert(
      'Create Settlement',
      `Send ${formatCurrency(settlement.amount)} to ${settlement.to_user_name || 'user'}?`,
      [
        {text: 'Cancel', style: 'cancel'},
        {
          text: 'Create Transaction',
          onPress: async () => {
            setSettling(settlement.to_user_id);
            try {
              await transactionService.create({
                group_id: groupId,
                to_user_id: settlement.to_user_id,
                amount: settlement.amount,
                note: `Settlement in ${groupName}`,
              });
              Alert.alert(
                'Success',
                'Transaction created! Waiting for confirmation.',
              );
              await loadData();
            } catch (error: any) {
              Alert.alert(
                'Error',
                error.message || 'Failed to create transaction',
              );
            } finally {
              setSettling(null);
            }
          },
        },
      ],
    );
  };

  const handlePayViaBank = (settlement: Settlement) => {
    navigation.navigate('Payment', {
      toUserId: settlement.to_user_id,
      toUserName: settlement.to_user_name || 'Unknown',
      amount: settlement.amount,
      groupId,
      groupName,
    });
  };

  const getBalanceColor = (amount: number): string => {
    if (amount > 0) return colors.success;
    if (amount < 0) return colors.error;
    return colors.textSecondary;
  };

  const getBalanceIcon = (amount: number): string => {
    if (amount > 0) return 'arrow-down-circle';
    if (amount < 0) return 'arrow-up-circle';
    return 'check-circle';
  };

  const getBalanceLabel = (amount: number): string => {
    if (amount > 0) return 'is owed';
    if (amount < 0) return 'owes';
    return 'settled';
  };

  const renderBalancesTab = () => (
    <View>
      {/* Summary Card */}
      <View style={styles.summaryCard}>
        <Text style={styles.summaryTitle}>Group Balance Summary</Text>
        <View style={styles.summaryStats}>
          <View style={styles.statItem}>
            <Icon name="trending-up" size={24} color={colors.success} />
            <Text style={styles.statValue}>
              {formatCurrency(
                balances
                  ?.filter((b: Balance) => b.net_balance > 0)
                  .reduce(
                    (sum: number, b: Balance) => sum + b.net_balance,
                    0,
                  ) || 0,
              )}
            </Text>
            <Text style={styles.statLabel}>Total Owed</Text>
          </View>
          <View style={styles.statDivider} />
          <View style={styles.statItem}>
            <Icon name="trending-down" size={24} color={colors.error} />
            <Text style={styles.statValue}>
              {formatCurrency(
                Math.abs(
                  balances
                    ?.filter((b: Balance) => b.net_balance < 0)
                    .reduce(
                      (sum: number, b: Balance) => sum + b.net_balance,
                      0,
                    ) || 0,
                ),
              )}
            </Text>
            <Text style={styles.statLabel}>Total Owes</Text>
          </View>
        </View>
      </View>

      {/* Balance List */}
      {balances && balances.length > 0 ? (
        balances.map((balance: Balance, index: number) => (
          <View key={index} style={styles.balanceCard}>
            <View style={styles.balanceLeft}>
              <View
                style={[
                  styles.balanceAvatar,
                  {
                    backgroundColor:
                      balance.net_balance >= 0
                        ? `${colors.success}20`
                        : `${colors.error}20`,
                  },
                ]}>
                <Icon
                  name={getBalanceIcon(balance.net_balance)}
                  size={24}
                  color={getBalanceColor(balance.net_balance)}
                />
              </View>
              <View>
                <Text style={styles.balanceName}>
                  {balance.user_name || 'Unknown'}
                  {balance.user_id === user?.id ? ' (You)' : ''}
                </Text>
                <Text
                  style={[
                    styles.balanceStatus,
                    {color: getBalanceColor(balance.net_balance)},
                  ]}>
                  {getBalanceLabel(balance.net_balance)}
                </Text>
              </View>
            </View>
            <View style={styles.balanceRight}>
              <Text
                style={[
                  styles.balanceAmount,
                  {color: getBalanceColor(balance.net_balance)},
                ]}>
                {balance.net_balance > 0 ? '+' : ''}
                {formatCurrency(balance.net_balance)}
              </Text>
              <View style={styles.balanceDetails}>
                <Text style={styles.balanceDetailText}>
                  Paid: {formatCurrency(balance.total_paid)}
                </Text>
                <Text style={styles.balanceDetailText}>
                  Share: {formatCurrency(balance.total_owed)}
                </Text>
              </View>
            </View>
          </View>
        ))
      ) : (
        <View style={styles.emptyState}>
          <Icon name="scale-balance" size={48} color={colors.textSecondary} />
          <Text style={styles.emptyText}>No balances yet</Text>
          <Text style={styles.emptySubtext}>
            Add bills to see balance calculations
          </Text>
        </View>
      )}
    </View>
  );

  const renderSettlementsTab = () => (
    <View>
      {/* Info Banner */}
      <View style={styles.infoBanner}>
        <Icon name="lightbulb-outline" size={20} color={colors.warning} />
        <Text style={styles.infoText}>
          Optimized settlements use the minimum number of transactions to settle
          all debts.
        </Text>
      </View>

      {/* Settlements List */}
      {settlements && settlements.length > 0 ? (
        settlements.map((settlement: Settlement, index: number) => (
          <View key={index} style={styles.settlementCard}>
            <View style={styles.settlementHeader}>
              <View style={styles.settlementFlow}>
                <View style={styles.userChip}>
                  <Text style={styles.userChipText}>
                    {(settlement.from_user_name || 'U')[0].toUpperCase()}
                  </Text>
                </View>
                <View style={styles.flowArrow}>
                  <Icon
                    name="arrow-right"
                    size={20}
                    color={colors.primary}
                  />
                </View>
                <View
                  style={[styles.userChip, {backgroundColor: colors.success}]}>
                  <Text style={styles.userChipText}>
                    {(settlement.to_user_name || 'U')[0].toUpperCase()}
                  </Text>
                </View>
              </View>
              <Text style={styles.settlementAmount}>
                {formatCurrency(settlement.amount)}
              </Text>
            </View>

            <View style={styles.settlementNames}>
              <Text style={styles.settlementFrom}>
                {settlement.from_user_name || 'Unknown'}
                {settlement.from_user_id === user?.id ? ' (You)' : ''}
              </Text>
              <Text style={styles.settlementTo}>
                {settlement.to_user_name || 'Unknown'}
                {settlement.to_user_id === user?.id ? ' (You)' : ''}
              </Text>
            </View>

            {/* Actions for current user */}
            {settlement.from_user_id === user?.id && (
              <View style={styles.settlementActions}>
                <TouchableOpacity
                  style={styles.settleButton}
                  onPress={() => handleSettle(settlement)}
                  disabled={settling === settlement.to_user_id}>
                  {settling === settlement.to_user_id ? (
                    <ActivityIndicator size="small" color="#FFFFFF" />
                  ) : (
                    <>
                      <Icon
                        name="check-circle"
                        size={16}
                        color="#FFFFFF"
                      />
                      <Text style={styles.settleButtonText}>
                        Mark as Paid
                      </Text>
                    </>
                  )}
                </TouchableOpacity>
                <TouchableOpacity
                  style={styles.bankButton}
                  onPress={() => handlePayViaBank(settlement)}>
                  <Icon
                    name="bank-transfer"
                    size={16}
                    color={colors.primary}
                  />
                  <Text style={styles.bankButtonText}>Pay via App</Text>
                </TouchableOpacity>
              </View>
            )}
          </View>
        ))
      ) : (
        <View style={styles.emptyState}>
          <Icon name="check-decagram" size={48} color={colors.success} />
          <Text style={styles.emptyText}>All settled up! 🎉</Text>
          <Text style={styles.emptySubtext}>
            No pending settlements in this group
          </Text>
        </View>
      )}
    </View>
  );

  if (loading && !refreshing) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={colors.primary} />
        <Text style={styles.loadingText}>Calculating balances...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Tab Switcher */}
      <View style={styles.tabContainer}>
        <TouchableOpacity
          style={[
            styles.tab,
            activeTab === 'balances' && styles.activeTab,
          ]}
          onPress={() => setActiveTab('balances')}>
          <Icon
            name="scale-balance"
            size={18}
            color={
              activeTab === 'balances' ? colors.primary : colors.textSecondary
            }
          />
          <Text
            style={[
              styles.tabText,
              activeTab === 'balances' && styles.activeTabText,
            ]}>
            Balances
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[
            styles.tab,
            activeTab === 'settlements' && styles.activeTab,
          ]}
          onPress={() => setActiveTab('settlements')}>
          <Icon
            name="swap-horizontal"
            size={18}
            color={
              activeTab === 'settlements'
                ? colors.primary
                : colors.textSecondary
            }
          />
          <Text
            style={[
              styles.tabText,
              activeTab === 'settlements' && styles.activeTabText,
            ]}>
            Settlements
          </Text>
        </TouchableOpacity>
      </View>

      <ScrollView
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
          />
        }>
        {activeTab === 'balances'
          ? renderBalancesTab()
          : renderSettlementsTab()}
        <View style={{height: spacing.xl}} />
      </ScrollView>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: colors.background,
  },
  loadingText: {
    marginTop: spacing.sm,
    fontSize: fontSize.md,
    color: colors.textSecondary,
  },
  tabContainer: {
    flexDirection: 'row',
    backgroundColor: colors.surface,
    margin: spacing.md,
    borderRadius: borderRadius.lg,
    padding: spacing.xs,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  tab: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.xs,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.md,
  },
  activeTab: {
    backgroundColor: colors.primaryLight,
  },
  tabText: {
    fontSize: fontSize.sm,
    fontWeight: '500',
    color: colors.textSecondary,
  },
  activeTabText: {
    color: colors.primary,
    fontWeight: '700',
  },
  summaryCard: {
    backgroundColor: colors.surface,
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  summaryTitle: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    textAlign: 'center',
    marginBottom: spacing.md,
  },
  summaryStats: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statItem: {
    flex: 1,
    alignItems: 'center',
  },
  statValue: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    marginTop: spacing.xs,
  },
  statLabel: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginTop: 2,
  },
  statDivider: {
    width: 1,
    height: 50,
    backgroundColor: colors.border,
  },
  balanceCard: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: colors.surface,
    marginHorizontal: spacing.md,
    marginBottom: spacing.sm,
    padding: spacing.md,
    borderRadius: borderRadius.lg,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  balanceLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  balanceAvatar: {
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
  },
  balanceName: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
  },
  balanceStatus: {
    fontSize: fontSize.xs,
    fontWeight: '500',
    marginTop: 2,
  },
  balanceRight: {
    alignItems: 'flex-end',
  },
  balanceAmount: {
    fontSize: fontSize.lg,
    fontWeight: '700',
  },
  balanceDetails: {
    marginTop: 2,
  },
  balanceDetailText: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    textAlign: 'right',
  },
  infoBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    backgroundColor: `${colors.warning}15`,
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    padding: spacing.md,
    borderRadius: borderRadius.md,
    borderLeftWidth: 3,
    borderLeftColor: colors.warning,
  },
  infoText: {
    flex: 1,
    fontSize: fontSize.sm,
    color: colors.text,
    lineHeight: 20,
  },
  settlementCard: {
    backgroundColor: colors.surface,
    marginHorizontal: spacing.md,
    marginBottom: spacing.sm,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  settlementHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  settlementFlow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  userChip: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
  },
  userChipText: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: '#FFFFFF',
  },
  flowArrow: {
    backgroundColor: colors.primaryLight,
    padding: spacing.xs,
    borderRadius: borderRadius.full,
  },
  settlementAmount: {
    fontSize: fontSize.xl,
    fontWeight: '800',
    color: colors.text,
  },
  settlementNames: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: spacing.md,
  },
  settlementFrom: {
    fontSize: fontSize.sm,
    color: colors.error,
    fontWeight: '500',
  },
  settlementTo: {
    fontSize: fontSize.sm,
    color: colors.success,
    fontWeight: '500',
  },
  settlementActions: {
    flexDirection: 'row',
    gap: spacing.sm,
  },
  settleButton: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.xs,
    backgroundColor: colors.primary,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.md,
  },
  settleButtonText: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  bankButton: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.xs,
    backgroundColor: colors.primaryLight,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.md,
  },
  bankButtonText: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.primary,
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: spacing.xxl,
  },
  emptyText: {
    fontSize: fontSize.lg,
    fontWeight: '600',
    color: colors.text,
    marginTop: spacing.md,
  },
  emptySubtext: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginTop: spacing.xs,
  },
});

export default BalancesScreen;
