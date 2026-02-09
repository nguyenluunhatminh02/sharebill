import React, {useEffect, useState} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  Share,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialCommunityIcons';
import {colors, spacing, borderRadius, fontSize} from '../../theme';
import {useBillStore} from '../../store/useBillStore';
import {useAuthStore} from '../../store/useAuthStore';
import type {Bill, BillSplit} from '../../types';

interface BillDetailScreenProps {
  navigation: any;
  route: {
    params: {
      billId: string;
      groupId: string;
    };
  };
}

const BillDetailScreen: React.FC<BillDetailScreenProps> = ({
  navigation,
  route,
}) => {
  const {billId, groupId} = route.params;
  const {currentBill, loading, fetchBillDetail, deleteBill} = useBillStore();
  const {user} = useAuthStore();
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    fetchBillDetail(billId);
  }, [billId]);

  const handleDelete = () => {
    Alert.alert(
      'Delete Bill',
      'Are you sure you want to delete this bill? This action cannot be undone.',
      [
        {text: 'Cancel', style: 'cancel'},
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            setDeleting(true);
            try {
              await deleteBill(billId);
              navigation.goBack();
            } catch (error: any) {
              Alert.alert('Error', error.message || 'Failed to delete bill');
            } finally {
              setDeleting(false);
            }
          },
        },
      ],
    );
  };

  const handleShare = async () => {
    if (!currentBill) return;

    let message = `💰 ${currentBill.title}\n`;
    message += `Total: ${formatCurrency(currentBill.total_amount)}\n\n`;
    message += `Split Details:\n`;

    currentBill.splits?.forEach(split => {
      message += `• ${split.user_name || split.user_id}: ${formatCurrency(split.amount)}\n`;
    });

    try {
      await Share.share({message});
    } catch (error) {
      console.error('Share error:', error);
    }
  };

  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('vi-VN', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getSplitTypeLabel = (type: string): string => {
    switch (type) {
      case 'equal':
        return 'Equal Split';
      case 'by_item':
        return 'By Item';
      case 'by_percentage':
        return 'By Percentage';
      case 'by_amount':
        return 'By Amount';
      default:
        return type;
    }
  };

  const getSplitTypeIcon = (type: string): string => {
    switch (type) {
      case 'equal':
        return 'equal';
      case 'by_item':
        return 'format-list-bulleted';
      case 'by_percentage':
        return 'percent';
      case 'by_amount':
        return 'currency-usd';
      default:
        return 'help-circle';
    }
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'active':
        return colors.success;
      case 'settled':
        return colors.primary;
      case 'cancelled':
        return colors.error;
      default:
        return colors.textSecondary;
    }
  };

  if (loading || !currentBill) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={colors.primary} />
        <Text style={styles.loadingText}>Loading bill details...</Text>
      </View>
    );
  }

  const isCreator = currentBill.created_by === user?.id;
  const myShare = currentBill.splits?.find(
    (s: BillSplit) => s.user_id === user?.id,
  );

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Header Card */}
      <View style={styles.headerCard}>
        <View style={styles.headerTop}>
          <View style={styles.statusBadge}>
            <View
              style={[
                styles.statusDot,
                {backgroundColor: getStatusColor(currentBill.status)},
              ]}
            />
            <Text
              style={[
                styles.statusText,
                {color: getStatusColor(currentBill.status)},
              ]}>
              {currentBill.status.toUpperCase()}
            </Text>
          </View>
          <TouchableOpacity onPress={handleShare}>
            <Icon name="share-variant" size={24} color={colors.primary} />
          </TouchableOpacity>
        </View>

        <Text style={styles.billTitle}>{currentBill.title}</Text>
        {currentBill.description ? (
          <Text style={styles.billDescription}>{currentBill.description}</Text>
        ) : null}

        <View style={styles.totalContainer}>
          <Text style={styles.totalLabel}>Total Amount</Text>
          <Text style={styles.totalAmount}>
            {formatCurrency(currentBill.total_amount)}
          </Text>
        </View>

        <View style={styles.metaRow}>
          <View style={styles.metaItem}>
            <Icon name="calendar" size={16} color={colors.textSecondary} />
            <Text style={styles.metaText}>
              {formatDate(currentBill.created_at)}
            </Text>
          </View>
          <View style={styles.metaItem}>
            <Icon
              name={getSplitTypeIcon(currentBill.split_type)}
              size={16}
              color={colors.textSecondary}
            />
            <Text style={styles.metaText}>
              {getSplitTypeLabel(currentBill.split_type)}
            </Text>
          </View>
        </View>
      </View>

      {/* My Share */}
      {myShare && (
        <View style={styles.myShareCard}>
          <View style={styles.myShareHeader}>
            <Icon name="account-circle" size={24} color={colors.primary} />
            <Text style={styles.myShareTitle}>Your Share</Text>
          </View>
          <Text style={styles.myShareAmount}>
            {formatCurrency(myShare.amount)}
          </Text>
          {myShare.items && myShare.items.length > 0 && (
            <View style={styles.myItems}>
              {myShare.items.map((item, index) => (
                <View key={index} style={styles.myItemRow}>
                  <Text style={styles.myItemName}>{item}</Text>
                </View>
              ))}
            </View>
          )}
        </View>
      )}

      {/* Items Section */}
      {currentBill.items && currentBill.items.length > 0 && (
        <View style={styles.sectionCard}>
          <View style={styles.sectionHeader}>
            <Icon
              name="format-list-bulleted"
              size={20}
              color={colors.primary}
            />
            <Text style={styles.sectionTitle}>Items</Text>
          </View>
          {currentBill.items.map((item, index) => (
            <View key={index} style={styles.itemRow}>
              <View style={styles.itemInfo}>
                <Text style={styles.itemName}>{item.name}</Text>
                <Text style={styles.itemQuantity}>
                  x{item.quantity} @ {formatCurrency(item.price)}
                </Text>
              </View>
              <Text style={styles.itemTotal}>
                {formatCurrency(item.quantity * item.price)}
              </Text>
            </View>
          ))}

          {/* Extra Charges */}
          {currentBill.extra_charges && (
            <View style={styles.extraCharges}>
              <View style={styles.divider} />
              {currentBill.extra_charges.tax > 0 && (
                <View style={styles.extraRow}>
                  <Text style={styles.extraLabel}>Tax</Text>
                  <Text style={styles.extraValue}>
                    {formatCurrency(currentBill.extra_charges.tax)}
                  </Text>
                </View>
              )}
              {currentBill.extra_charges.service_charge > 0 && (
                <View style={styles.extraRow}>
                  <Text style={styles.extraLabel}>Service Charge</Text>
                  <Text style={styles.extraValue}>
                    {formatCurrency(currentBill.extra_charges.service_charge)}
                  </Text>
                </View>
              )}
              {currentBill.extra_charges.discount > 0 && (
                <View style={styles.extraRow}>
                  <Text style={[styles.extraLabel, {color: colors.success}]}>
                    Discount
                  </Text>
                  <Text style={[styles.extraValue, {color: colors.success}]}>
                    -{formatCurrency(currentBill.extra_charges.discount)}
                  </Text>
                </View>
              )}
            </View>
          )}
        </View>
      )}

      {/* Split Details */}
      <View style={styles.sectionCard}>
        <View style={styles.sectionHeader}>
          <Icon name="account-group" size={20} color={colors.primary} />
          <Text style={styles.sectionTitle}>Split Details</Text>
          <Text style={styles.splitCount}>
            {currentBill.splits?.length || 0} people
          </Text>
        </View>
        {currentBill.splits?.map((split: BillSplit, index: number) => (
          <View key={index} style={styles.splitRow}>
            <View style={styles.splitAvatar}>
              <Text style={styles.splitAvatarText}>
                {(split.user_name || 'U')[0].toUpperCase()}
              </Text>
            </View>
            <View style={styles.splitInfo}>
              <Text style={styles.splitName}>
                {split.user_name || 'Unknown'}
                {split.user_id === user?.id ? ' (You)' : ''}
              </Text>
              {split.items && split.items.length > 0 && (
                <Text style={styles.splitItems} numberOfLines={1}>
                  {split.items.join(', ')}
                </Text>
              )}
            </View>
            <View style={styles.splitAmountContainer}>
              <Text style={styles.splitAmount}>
                {formatCurrency(split.amount)}
              </Text>
              {split.percentage ? (
                <Text style={styles.splitPercentage}>
                  {split.percentage.toFixed(1)}%
                </Text>
              ) : null}
            </View>
          </View>
        ))}
      </View>

      {/* Actions */}
      {isCreator && currentBill.status === 'active' && (
        <View style={styles.actionsCard}>
          <TouchableOpacity
            style={styles.deleteButton}
            onPress={handleDelete}
            disabled={deleting}>
            {deleting ? (
              <ActivityIndicator size="small" color={colors.error} />
            ) : (
              <>
                <Icon name="delete-outline" size={20} color={colors.error} />
                <Text style={styles.deleteButtonText}>Delete Bill</Text>
              </>
            )}
          </TouchableOpacity>
        </View>
      )}

      <View style={{height: spacing.xl}} />
    </ScrollView>
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
  headerCard: {
    backgroundColor: colors.surface,
    margin: spacing.md,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  headerTop: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.background,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    borderRadius: borderRadius.full,
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: spacing.xs,
  },
  statusText: {
    fontSize: fontSize.xs,
    fontWeight: '700',
  },
  billTitle: {
    fontSize: fontSize.xxl,
    fontWeight: '700',
    color: colors.text,
    marginBottom: spacing.xs,
  },
  billDescription: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginBottom: spacing.md,
  },
  totalContainer: {
    backgroundColor: colors.primaryLight,
    padding: spacing.md,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  totalLabel: {
    fontSize: fontSize.sm,
    color: colors.primary,
    fontWeight: '500',
  },
  totalAmount: {
    fontSize: 32,
    fontWeight: '800',
    color: colors.primary,
    marginTop: spacing.xs,
  },
  metaRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  metaItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
  },
  metaText: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
  },
  myShareCard: {
    backgroundColor: colors.primary,
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
  },
  myShareHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    marginBottom: spacing.sm,
  },
  myShareTitle: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: '#FFFFFF',
  },
  myShareAmount: {
    fontSize: 28,
    fontWeight: '800',
    color: '#FFFFFF',
  },
  myItems: {
    marginTop: spacing.sm,
  },
  myItemRow: {
    paddingVertical: spacing.xs,
  },
  myItemName: {
    fontSize: fontSize.sm,
    color: 'rgba(255,255,255,0.8)',
  },
  sectionCard: {
    backgroundColor: colors.surface,
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    marginBottom: spacing.md,
  },
  sectionTitle: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    flex: 1,
  },
  splitCount: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
  },
  itemRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  itemInfo: {
    flex: 1,
  },
  itemName: {
    fontSize: fontSize.md,
    fontWeight: '500',
    color: colors.text,
  },
  itemQuantity: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginTop: 2,
  },
  itemTotal: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
  },
  extraCharges: {
    marginTop: spacing.sm,
  },
  divider: {
    height: 1,
    backgroundColor: colors.border,
    marginVertical: spacing.sm,
  },
  extraRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: spacing.xs,
  },
  extraLabel: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
  },
  extraValue: {
    fontSize: fontSize.sm,
    fontWeight: '500',
    color: colors.text,
  },
  splitRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  splitAvatar: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: colors.primaryLight,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.sm,
  },
  splitAvatarText: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.primary,
  },
  splitInfo: {
    flex: 1,
  },
  splitName: {
    fontSize: fontSize.md,
    fontWeight: '500',
    color: colors.text,
  },
  splitItems: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginTop: 2,
  },
  splitAmountContainer: {
    alignItems: 'flex-end',
  },
  splitAmount: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.text,
  },
  splitPercentage: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginTop: 2,
  },
  actionsCard: {
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
  },
  deleteButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
    backgroundColor: colors.surface,
    padding: spacing.md,
    borderRadius: borderRadius.lg,
    borderWidth: 1,
    borderColor: colors.error,
  },
  deleteButtonText: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.error,
  },
});

export default BillDetailScreen;
