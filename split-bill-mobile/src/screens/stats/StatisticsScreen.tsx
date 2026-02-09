import React, {useEffect, useState, useCallback} from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  RefreshControl,
  TouchableOpacity,
  Share,
  ActivityIndicator,
  Dimensions,
} from 'react-native';
import {useRoute, RouteProp} from '@react-navigation/native';
import Icon from 'react-native-vector-icons/Ionicons';
import {colors, spacing, fontSize, borderRadius} from '../../theme';
import {statsAPI} from '../../api/services';
import {
  GroupStats,
  CategoryStat,
  MemberSpendStats,
  MonthlySpend,
} from '../../types';
import {RootStackParamList} from '../../navigation/AppNavigator';

type RouteProps = RouteProp<RootStackParamList, 'Statistics'>;

const SCREEN_WIDTH = Dimensions.get('window').width;

export default function StatisticsScreen() {
  const route = useRoute<RouteProps>();
  const {groupId, groupName} = route.params;

  const [stats, setStats] = useState<GroupStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState<'overview' | 'categories' | 'members'>('overview');

  const fetchStats = async () => {
    try {
      const res = await statsAPI.getGroupStats(groupId);
      if (res.data?.data) {
        setStats(res.data.data);
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchStats();
    setRefreshing(false);
  }, []);

  const handleExport = async () => {
    try {
      const res = await statsAPI.exportGroupSummary(groupId, 'json');
      const summary = res.data?.data?.summary || '';
      await Share.share({
        message: summary,
        title: `Tổng kết nhóm ${groupName}`,
      });
    } catch (error) {
      console.error('Export failed:', error);
    }
  };

  const formatCurrency = (amount: number): string => {
    if (amount >= 1000000) {
      return `${(amount / 1000000).toFixed(1)}tr₫`;
    }
    if (amount >= 1000) {
      return `${(amount / 1000).toFixed(0)}k₫`;
    }
    return `${amount.toFixed(0)}₫`;
  };

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color={colors.primary} />
          <Text style={styles.loadingText}>Đang tải thống kê...</Text>
        </View>
      </SafeAreaView>
    );
  }

  if (!stats) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.emptyState}>
          <Icon name="bar-chart-outline" size={64} color={colors.textSecondary} />
          <Text style={styles.emptyText}>Không có dữ liệu thống kê</Text>
        </View>
      </SafeAreaView>
    );
  }

  const maxMonthly = Math.max(...(stats.monthly_trend?.map(m => m.total) || [1]));

  const renderOverview = () => (
    <View>
      {/* Summary Cards */}
      <View style={styles.summaryGrid}>
        <View style={[styles.summaryCard, {backgroundColor: '#FF6B6B15'}]}>
          <Icon name="cash-outline" size={28} color="#FF6B6B" />
          <Text style={styles.summaryValue}>{formatCurrency(stats.total_spent)}</Text>
          <Text style={styles.summaryLabel}>Tổng chi tiêu</Text>
        </View>
        <View style={[styles.summaryCard, {backgroundColor: '#1E90FF15'}]}>
          <Icon name="receipt-outline" size={28} color="#1E90FF" />
          <Text style={styles.summaryValue}>{stats.total_bills}</Text>
          <Text style={styles.summaryLabel}>Hóa đơn</Text>
        </View>
        <View style={[styles.summaryCard, {backgroundColor: '#2ED57315'}]}>
          <Icon name="people-outline" size={28} color="#2ED573" />
          <Text style={styles.summaryValue}>{stats.total_members}</Text>
          <Text style={styles.summaryLabel}>Thành viên</Text>
        </View>
        <View style={[styles.summaryCard, {backgroundColor: '#A29BFE15'}]}>
          <Icon name="calculator-outline" size={28} color="#A29BFE" />
          <Text style={styles.summaryValue}>{formatCurrency(stats.average_bill)}</Text>
          <Text style={styles.summaryLabel}>TB/hóa đơn</Text>
        </View>
      </View>

      {/* Largest & Smallest Bill */}
      {stats.largest_bill && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>📊 Hóa đơn nổi bật</Text>
          <View style={styles.highlightRow}>
            <View style={styles.highlightCard}>
              <Icon name="arrow-up-circle" size={20} color="#FF6B6B" />
              <Text style={styles.highlightLabel}>Lớn nhất</Text>
              <Text style={styles.highlightValue}>{formatCurrency(stats.largest_bill.amount)}</Text>
              <Text style={styles.highlightTitle} numberOfLines={1}>{stats.largest_bill.title}</Text>
            </View>
            {stats.smallest_bill && (
              <View style={styles.highlightCard}>
                <Icon name="arrow-down-circle" size={20} color="#2ED573" />
                <Text style={styles.highlightLabel}>Nhỏ nhất</Text>
                <Text style={styles.highlightValue}>{formatCurrency(stats.smallest_bill.amount)}</Text>
                <Text style={styles.highlightTitle} numberOfLines={1}>{stats.smallest_bill.title}</Text>
              </View>
            )}
          </View>
        </View>
      )}

      {/* Monthly Trend - Simple Bar Chart */}
      {stats.monthly_trend && stats.monthly_trend.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>📈 Chi tiêu theo tháng</Text>
          <View style={styles.chartContainer}>
            {stats.monthly_trend.map((month: MonthlySpend, index: number) => (
              <View key={index} style={styles.barGroup}>
                <Text style={styles.barValue}>{formatCurrency(month.total)}</Text>
                <View style={styles.barWrapper}>
                  <View
                    style={[
                      styles.bar,
                      {
                        height: maxMonthly > 0 ? (month.total / maxMonthly) * 120 : 0,
                        backgroundColor: colors.primary,
                      },
                    ]}
                  />
                </View>
                <Text style={styles.barLabel}>{month.month}</Text>
              </View>
            ))}
          </View>
        </View>
      )}

      {/* Recent Bills */}
      {stats.recent_bills && stats.recent_bills.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>🕐 Hóa đơn gần đây</Text>
          {stats.recent_bills.map((bill, index) => (
            <View key={index} style={styles.recentBillRow}>
              <View style={styles.recentBillInfo}>
                <Text style={styles.recentBillTitle}>{bill.title}</Text>
                <Text style={styles.recentBillMeta}>
                  {bill.paid_by_name} • {new Date(bill.created_at).toLocaleDateString('vi-VN')}
                </Text>
              </View>
              <Text style={styles.recentBillAmount}>{formatCurrency(bill.amount)}</Text>
            </View>
          ))}
        </View>
      )}
    </View>
  );

  const renderCategories = () => (
    <View>
      {stats.category_stats && stats.category_stats.length > 0 ? (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>📁 Chi tiêu theo danh mục</Text>
          {stats.category_stats.map((cat: CategoryStat, index: number) => (
            <View key={index} style={styles.categoryRow}>
              <View style={[styles.categoryIcon, {backgroundColor: cat.color + '20'}]}>
                <Icon name={cat.icon || 'ellipsis-horizontal'} size={20} color={cat.color} />
              </View>
              <View style={styles.categoryInfo}>
                <View style={styles.categoryHeader}>
                  <Text style={styles.categoryName}>{cat.category}</Text>
                  <Text style={styles.categoryAmount}>{formatCurrency(cat.total)}</Text>
                </View>
                <View style={styles.progressBarBg}>
                  <View
                    style={[
                      styles.progressBarFill,
                      {width: `${Math.min(cat.percentage, 100)}%`, backgroundColor: cat.color},
                    ]}
                  />
                </View>
                <Text style={styles.categoryMeta}>
                  {cat.count} hóa đơn • {cat.percentage.toFixed(1)}%
                </Text>
              </View>
            </View>
          ))}
        </View>
      ) : (
        <View style={styles.emptyState}>
          <Icon name="pie-chart-outline" size={48} color={colors.textSecondary} />
          <Text style={styles.emptyText}>Chưa có dữ liệu danh mục</Text>
        </View>
      )}
    </View>
  );

  const renderMembers = () => (
    <View>
      {stats.member_stats && stats.member_stats.length > 0 ? (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>👥 Chi tiêu theo thành viên</Text>
          {stats.member_stats.map((member: MemberSpendStats, index: number) => (
            <View key={index} style={styles.memberCard}>
              <View style={styles.memberHeader}>
                <View style={styles.memberAvatar}>
                  <Text style={styles.memberAvatarText}>
                    {(member.display_name || '?')[0].toUpperCase()}
                  </Text>
                </View>
                <View style={styles.memberInfo}>
                  <Text style={styles.memberName}>{member.display_name}</Text>
                  <Text style={styles.memberBills}>{member.bill_count} hóa đơn</Text>
                </View>
                <Text style={styles.memberPercentage}>{member.percentage.toFixed(0)}%</Text>
              </View>
              <View style={styles.memberStats}>
                <View style={styles.memberStatItem}>
                  <Text style={styles.memberStatLabel}>Đã trả</Text>
                  <Text style={[styles.memberStatValue, {color: '#2ED573'}]}>
                    {formatCurrency(member.total_paid)}
                  </Text>
                </View>
                <View style={styles.memberStatItem}>
                  <Text style={styles.memberStatLabel}>Phần phải trả</Text>
                  <Text style={[styles.memberStatValue, {color: '#FF6B6B'}]}>
                    {formatCurrency(member.total_owed)}
                  </Text>
                </View>
                <View style={styles.memberStatItem}>
                  <Text style={styles.memberStatLabel}>Số dư</Text>
                  <Text
                    style={[
                      styles.memberStatValue,
                      {color: member.net_balance >= 0 ? '#2ED573' : '#FF6B6B'},
                    ]}>
                    {member.net_balance >= 0 ? '+' : ''}
                    {formatCurrency(member.net_balance)}
                  </Text>
                </View>
              </View>
            </View>
          ))}
        </View>
      ) : (
        <View style={styles.emptyState}>
          <Icon name="people-outline" size={48} color={colors.textSecondary} />
          <Text style={styles.emptyText}>Chưa có dữ liệu thành viên</Text>
        </View>
      )}
    </View>
  );

  return (
    <SafeAreaView style={styles.container}>
      {/* Tab Bar */}
      <View style={styles.tabBar}>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'overview' && styles.tabActive]}
          onPress={() => setActiveTab('overview')}>
          <Icon
            name="bar-chart"
            size={18}
            color={activeTab === 'overview' ? colors.primary : colors.textSecondary}
          />
          <Text style={[styles.tabText, activeTab === 'overview' && styles.tabTextActive]}>
            Tổng quan
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'categories' && styles.tabActive]}
          onPress={() => setActiveTab('categories')}>
          <Icon
            name="pie-chart"
            size={18}
            color={activeTab === 'categories' ? colors.primary : colors.textSecondary}
          />
          <Text style={[styles.tabText, activeTab === 'categories' && styles.tabTextActive]}>
            Danh mục
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'members' && styles.tabActive]}
          onPress={() => setActiveTab('members')}>
          <Icon
            name="people"
            size={18}
            color={activeTab === 'members' ? colors.primary : colors.textSecondary}
          />
          <Text style={[styles.tabText, activeTab === 'members' && styles.tabTextActive]}>
            Thành viên
          </Text>
        </TouchableOpacity>
      </View>

      <ScrollView
        style={styles.content}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}>
        {activeTab === 'overview' && renderOverview()}
        {activeTab === 'categories' && renderCategories()}
        {activeTab === 'members' && renderMembers()}
      </ScrollView>

      {/* Export FAB */}
      <TouchableOpacity style={styles.fab} onPress={handleExport}>
        <Icon name="share-outline" size={24} color="#FFF" />
      </TouchableOpacity>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, backgroundColor: colors.background},
  loadingContainer: {flex: 1, justifyContent: 'center', alignItems: 'center'},
  loadingText: {marginTop: spacing.md, color: colors.textSecondary, fontSize: fontSize.md},
  emptyState: {flex: 1, justifyContent: 'center', alignItems: 'center', padding: spacing.xl},
  emptyText: {marginTop: spacing.md, color: colors.textSecondary, fontSize: fontSize.md},

  // Tab Bar
  tabBar: {
    flexDirection: 'row',
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  tab: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: spacing.md,
    gap: 4,
  },
  tabActive: {
    borderBottomWidth: 2,
    borderBottomColor: colors.primary,
  },
  tabText: {fontSize: fontSize.sm, color: colors.textSecondary, fontWeight: '500'},
  tabTextActive: {color: colors.primary, fontWeight: '600'},

  content: {flex: 1, padding: spacing.md},

  // Summary Grid
  summaryGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginBottom: spacing.md,
  },
  summaryCard: {
    width: (SCREEN_WIDTH - spacing.md * 2 - spacing.sm) / 2 - 1,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    alignItems: 'center',
  },
  summaryValue: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    marginTop: spacing.xs,
  },
  summaryLabel: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginTop: 2,
  },

  // Sections
  section: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.md,
  },
  sectionTitle: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.text,
    marginBottom: spacing.md,
  },

  // Highlight cards
  highlightRow: {flexDirection: 'row', gap: spacing.sm},
  highlightCard: {
    flex: 1,
    backgroundColor: colors.background,
    borderRadius: borderRadius.sm,
    padding: spacing.md,
    alignItems: 'center',
  },
  highlightLabel: {fontSize: fontSize.xs, color: colors.textSecondary, marginTop: 4},
  highlightValue: {fontSize: fontSize.lg, fontWeight: '700', color: colors.text, marginTop: 2},
  highlightTitle: {fontSize: fontSize.xs, color: colors.textSecondary, marginTop: 2},

  // Bar Chart
  chartContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'flex-end',
    height: 180,
    paddingTop: spacing.md,
  },
  barGroup: {alignItems: 'center', flex: 1},
  barValue: {fontSize: 10, color: colors.textSecondary, marginBottom: 4},
  barWrapper: {height: 120, justifyContent: 'flex-end'},
  bar: {width: 24, borderRadius: 4, minHeight: 4},
  barLabel: {fontSize: 10, color: colors.textSecondary, marginTop: 4},

  // Recent Bills
  recentBillRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  recentBillInfo: {flex: 1},
  recentBillTitle: {fontSize: fontSize.md, fontWeight: '500', color: colors.text},
  recentBillMeta: {fontSize: fontSize.xs, color: colors.textSecondary, marginTop: 2},
  recentBillAmount: {fontSize: fontSize.md, fontWeight: '600', color: colors.primary},

  // Category Stats
  categoryRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  categoryIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  categoryInfo: {flex: 1},
  categoryHeader: {flexDirection: 'row', justifyContent: 'space-between', marginBottom: 4},
  categoryName: {fontSize: fontSize.md, fontWeight: '500', color: colors.text, textTransform: 'capitalize'},
  categoryAmount: {fontSize: fontSize.md, fontWeight: '600', color: colors.text},
  progressBarBg: {
    height: 6,
    backgroundColor: colors.border,
    borderRadius: 3,
    overflow: 'hidden',
    marginBottom: 4,
  },
  progressBarFill: {height: '100%', borderRadius: 3},
  categoryMeta: {fontSize: fontSize.xs, color: colors.textSecondary},

  // Member Stats
  memberCard: {
    backgroundColor: colors.background,
    borderRadius: borderRadius.sm,
    padding: spacing.md,
    marginBottom: spacing.sm,
  },
  memberHeader: {flexDirection: 'row', alignItems: 'center', marginBottom: spacing.sm},
  memberAvatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.sm,
  },
  memberAvatarText: {color: '#FFF', fontWeight: '700', fontSize: fontSize.md},
  memberInfo: {flex: 1},
  memberName: {fontSize: fontSize.md, fontWeight: '600', color: colors.text},
  memberBills: {fontSize: fontSize.xs, color: colors.textSecondary},
  memberPercentage: {fontSize: fontSize.lg, fontWeight: '700', color: colors.primary},
  memberStats: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    borderTopWidth: 1,
    borderTopColor: colors.border,
    paddingTop: spacing.sm,
  },
  memberStatItem: {alignItems: 'center'},
  memberStatLabel: {fontSize: fontSize.xs, color: colors.textSecondary, marginBottom: 2},
  memberStatValue: {fontSize: fontSize.sm, fontWeight: '600'},

  // FAB
  fab: {
    position: 'absolute',
    right: spacing.lg,
    bottom: spacing.xl,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
    elevation: 4,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.25,
    shadowRadius: 4,
  },
});
