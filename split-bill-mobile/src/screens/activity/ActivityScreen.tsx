import React, {useEffect, useState, useCallback} from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import Icon from 'react-native-vector-icons/Ionicons';
import {NativeStackScreenProps} from '@react-navigation/native-stack';
import {RootStackParamList} from '../../navigation/AppNavigator';
import {colors, spacing, borderRadius, fontSize, fontWeight} from '../../theme';
import {activityAPI} from '../../api/services';
import {Activity, ActivityType} from '../../types';

type Props = NativeStackScreenProps<RootStackParamList, 'Activity'>;

const getActivityIcon = (type: ActivityType): {name: string; color: string} => {
  switch (type) {
    case 'bill_created':
      return {name: 'receipt-outline', color: colors.primary};
    case 'bill_deleted':
      return {name: 'trash-outline', color: colors.error};
    case 'bill_updated':
      return {name: 'create-outline', color: colors.info};
    case 'member_joined':
      return {name: 'person-add-outline', color: colors.success};
    case 'member_left':
      return {name: 'person-remove-outline', color: colors.warning};
    case 'payment_sent':
      return {name: 'send-outline', color: colors.secondary};
    case 'payment_confirmed':
      return {name: 'checkmark-circle-outline', color: colors.success};
    case 'payment_rejected':
      return {name: 'close-circle-outline', color: colors.error};
    case 'group_created':
      return {name: 'people-outline', color: colors.primary};
    case 'settlement_created':
      return {name: 'swap-horizontal-outline', color: colors.accent};
    default:
      return {name: 'ellipse-outline', color: colors.textLight};
  }
};

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('vi-VN', {
    style: 'currency',
    currency: 'VND',
    minimumFractionDigits: 0,
  }).format(amount);
};

export default function ActivityScreen({route}: Props) {
  const {groupId, groupName} = route.params;

  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadActivities = useCallback(async () => {
    try {
      const res = groupId
        ? await activityAPI.getGroupActivities(groupId, 50)
        : await activityAPI.getMyActivities(50);

      setActivities(res.data?.data || []);
    } catch (error) {
      console.error('Failed to load activities:', error);
    } finally {
      setLoading(false);
    }
  }, [groupId]);

  useEffect(() => {
    loadActivities();
  }, [loadActivities]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadActivities();
    setRefreshing(false);
  }, [loadActivities]);

  const renderActivityItem = ({item}: {item: Activity}) => {
    const iconInfo = getActivityIcon(item.type);

    return (
      <View style={styles.activityItem}>
        <View
          style={[
            styles.activityIcon,
            {backgroundColor: iconInfo.color + '15'},
          ]}>
          <Icon name={iconInfo.name} size={22} color={iconInfo.color} />
        </View>
        <View style={styles.activityContent}>
          <View style={styles.activityHeader}>
            <Text style={styles.activityTitle} numberOfLines={1}>
              {item.title}
            </Text>
            <Text style={styles.activityTime}>{item.time_ago}</Text>
          </View>
          <Text style={styles.activityDetail} numberOfLines={2}>
            {item.detail}
          </Text>
          {item.amount !== undefined && item.amount > 0 && (
            <Text style={styles.activityAmount}>
              {formatCurrency(item.amount)}
            </Text>
          )}
          {item.group_name && !groupId && (
            <View style={styles.groupBadge}>
              <Icon
                name="people-outline"
                size={12}
                color={colors.textSecondary}
              />
              <Text style={styles.groupBadgeText}>{item.group_name}</Text>
            </View>
          )}
        </View>
      </View>
    );
  };

  const renderHeader = () => (
    <View style={styles.header}>
      <View style={styles.headerIcon}>
        <Icon name="time-outline" size={32} color={colors.primary} />
      </View>
      <Text style={styles.headerTitle}>
        {groupName ? `Hoạt động - ${groupName}` : 'Hoạt động gần đây'}
      </Text>
      <Text style={styles.headerSubtitle}>
        {activities.length > 0
          ? `${activities.length} hoạt động`
          : 'Chưa có hoạt động nào'}
      </Text>
    </View>
  );

  const renderEmpty = () => (
    <View style={styles.emptyState}>
      <Icon name="document-text-outline" size={64} color={colors.textLight} />
      <Text style={styles.emptyTitle}>Chưa có hoạt động</Text>
      <Text style={styles.emptySubtitle}>
        Các hoạt động trong nhóm sẽ xuất hiện ở đây
      </Text>
    </View>
  );

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={colors.primary} />
        <Text style={styles.loadingText}>Đang tải hoạt động...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <FlatList
        data={activities}
        renderItem={renderActivityItem}
        keyExtractor={item => item.id}
        ListHeaderComponent={renderHeader}
        ListEmptyComponent={renderEmpty}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        contentContainerStyle={
          activities.length === 0 ? styles.emptyContainer : undefined
        }
        showsVerticalScrollIndicator={false}
        ItemSeparatorComponent={() => <View style={styles.separator} />}
      />
    </View>
  );
}

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
    marginTop: spacing.md,
    color: colors.textSecondary,
    fontSize: fontSize.md,
  },
  emptyContainer: {
    flexGrow: 1,
  },
  header: {
    alignItems: 'center',
    paddingVertical: spacing.lg,
    paddingHorizontal: spacing.md,
  },
  headerIcon: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: colors.primaryLight + '30',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  headerTitle: {
    fontSize: fontSize.xl,
    fontWeight: fontWeight.bold,
    color: colors.text,
  },
  headerSubtitle: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    marginTop: spacing.xs,
  },
  activityItem: {
    flexDirection: 'row',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    backgroundColor: colors.surface,
  },
  activityIcon: {
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  activityContent: {
    flex: 1,
  },
  activityHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.xs,
  },
  activityTitle: {
    fontSize: fontSize.md,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    flex: 1,
    marginRight: spacing.sm,
  },
  activityTime: {
    fontSize: fontSize.xs,
    color: colors.textLight,
  },
  activityDetail: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    lineHeight: 18,
  },
  activityAmount: {
    fontSize: fontSize.md,
    fontWeight: fontWeight.bold,
    color: colors.primary,
    marginTop: spacing.xs,
  },
  groupBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: spacing.xs,
    backgroundColor: colors.background,
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
    borderRadius: borderRadius.sm,
    alignSelf: 'flex-start',
  },
  groupBadgeText: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginLeft: 4,
  },
  separator: {
    height: 1,
    backgroundColor: colors.borderLight,
    marginLeft: 72,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: spacing.xl,
    paddingVertical: spacing.xxl,
  },
  emptyTitle: {
    fontSize: fontSize.xl,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginTop: spacing.md,
  },
  emptySubtitle: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    textAlign: 'center',
    marginTop: spacing.sm,
  },
});
