import React, {useEffect, useState, useCallback} from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Image,
  Alert,
  Linking,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import Icon from 'react-native-vector-icons/Ionicons';
import {NativeStackScreenProps} from '@react-navigation/native-stack';
import {RootStackParamList} from '../../navigation/AppNavigator';
import {colors, spacing, borderRadius, fontSize, fontWeight} from '../../theme';
import {paymentAPI} from '../../api/services';
import {BankingDeeplink, BankInfo} from '../../types';

type Props = NativeStackScreenProps<RootStackParamList, 'Payment'>;

export default function PaymentScreen({route}: Props) {
  const {
    toUserId,
    toUserName,
    amount,
    groupId,
    groupName,
  } = route.params;

  const [deeplinks, setDeeplinks] = useState<BankingDeeplink[]>([]);
  const [vietQRUrl, setVietQRUrl] = useState<string>('');
  const [banks, setBanks] = useState<BankInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [selectedBank, setSelectedBank] = useState<BankInfo | null>(null);
  const [paymentInfo, setPaymentInfo] = useState<any>(null);

  const loadPaymentData = useCallback(async () => {
    try {
      // Load user payment info and supported banks in parallel
      const [paymentInfoRes, banksRes] = await Promise.all([
        paymentAPI.getUserPaymentInfo(toUserId),
        paymentAPI.getSupportedBanks(),
      ]);

      const userPayment = paymentInfoRes.data?.data;
      const bankList = banksRes.data?.data || [];

      setPaymentInfo(userPayment);
      setBanks(bankList);

      // If user has bank accounts, generate deeplinks
      if (userPayment?.bank_accounts?.length > 0) {
        const primaryAccount = userPayment.bank_accounts[0];
        const note = `Split Bill - ${groupName}`;

        const deeplinkRes = await paymentAPI.generateDeeplink({
          bank_code: primaryAccount.bank_code,
          account_number: primaryAccount.account_number,
          account_name: primaryAccount.account_name || toUserName,
          amount,
          note,
        });

        const deeplinkData = deeplinkRes.data?.data;
        if (deeplinkData) {
          setDeeplinks(deeplinkData.deeplinks || []);
          setVietQRUrl(deeplinkData.vietqr_url || '');
        }

        // Find matching bank
        const matchedBank = bankList.find(
          (b: BankInfo) => b.id === primaryAccount.bank_code || b.short_name === primaryAccount.bank_code,
        );
        if (matchedBank) {
          setSelectedBank(matchedBank);
        }
      }
    } catch (error) {
      console.error('Failed to load payment data:', error);
    } finally {
      setLoading(false);
    }
  }, [toUserId, toUserName, amount, groupName]);

  useEffect(() => {
    loadPaymentData();
  }, [loadPaymentData]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await loadPaymentData();
    setRefreshing(false);
  }, [loadPaymentData]);

  const formatCurrency = (value: number): string => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
      minimumFractionDigits: 0,
    }).format(value);
  };

  const handleOpenDeeplink = async (deeplink: BankingDeeplink) => {
    try {
      const canOpen = await Linking.canOpenURL(deeplink.scheme);
      if (canOpen) {
        await Linking.openURL(deeplink.scheme);
      } else {
        Alert.alert(
          'Không tìm thấy ứng dụng',
          `Ứng dụng ${deeplink.app_name} chưa được cài đặt trên thiết bị của bạn.`,
          [{text: 'OK'}],
        );
      }
    } catch (error) {
      Alert.alert('Lỗi', 'Không thể mở ứng dụng thanh toán.');
    }
  };

  const handleGenerateQR = async (bank: BankInfo) => {
    if (!paymentInfo?.bank_accounts?.length) {
      Alert.alert('Lỗi', 'Người nhận chưa cập nhật thông tin ngân hàng.');
      return;
    }

    const account = paymentInfo.bank_accounts[0];
    try {
      const res = await paymentAPI.generateVietQR({
        bank_id: bank.id,
        account_no: account.account_number,
        account_name: account.account_name || toUserName,
        amount,
        description: `Split Bill - ${groupName}`,
      });

      const data = res.data?.data;
      if (data?.qr_url) {
        setVietQRUrl(data.qr_url);
        setSelectedBank(bank);
      }
    } catch (error) {
      Alert.alert('Lỗi', 'Không thể tạo mã QR.');
    }
  };

  const getDeeplinkIcon = (iconName: string): string => {
    switch (iconName) {
      case 'wallet':
        return 'wallet-outline';
      case 'credit-card':
        return 'card-outline';
      case 'bank':
        return 'business-outline';
      default:
        return 'cash-outline';
    }
  };

  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={colors.primary} />
        <Text style={styles.loadingText}>Đang tải thông tin thanh toán...</Text>
      </View>
    );
  }

  return (
    <ScrollView
      style={styles.container}
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
      }>
      {/* Payment Summary */}
      <View style={styles.summaryCard}>
        <View style={styles.summaryHeader}>
          <Icon name="cash-outline" size={28} color={colors.primary} />
          <Text style={styles.summaryTitle}>Thanh toán</Text>
        </View>
        <Text style={styles.summaryAmount}>{formatCurrency(amount)}</Text>
        <View style={styles.summaryRow}>
          <Text style={styles.summaryLabel}>Cho:</Text>
          <Text style={styles.summaryValue}>{toUserName}</Text>
        </View>
        <View style={styles.summaryRow}>
          <Text style={styles.summaryLabel}>Nhóm:</Text>
          <Text style={styles.summaryValue}>{groupName}</Text>
        </View>
      </View>

      {/* VietQR Code */}
      {vietQRUrl ? (
        <View style={styles.qrSection}>
          <Text style={styles.sectionTitle}>
            <Icon name="qr-code-outline" size={20} color={colors.text} /> Mã QR Thanh Toán
          </Text>
          <View style={styles.qrCard}>
            {selectedBank && (
              <View style={styles.bankHeader}>
                <Image
                  source={{uri: selectedBank.logo}}
                  style={styles.bankLogo}
                  resizeMode="contain"
                />
                <Text style={styles.bankName}>{selectedBank.name}</Text>
              </View>
            )}
            <Image
              source={{uri: vietQRUrl}}
              style={styles.qrImage}
              resizeMode="contain"
            />
            {paymentInfo?.bank_accounts?.[0] && (
              <View style={styles.accountInfo}>
                <Text style={styles.accountNumber}>
                  STK: {paymentInfo.bank_accounts[0].account_number}
                </Text>
                <Text style={styles.accountName}>
                  {paymentInfo.bank_accounts[0].account_name || toUserName}
                </Text>
              </View>
            )}
            <Text style={styles.qrHint}>
              Quét mã QR bằng ứng dụng ngân hàng để thanh toán
            </Text>
          </View>
        </View>
      ) : (
        <View style={styles.noQRCard}>
          <Icon name="alert-circle-outline" size={40} color={colors.warning} />
          <Text style={styles.noQRText}>
            Người nhận chưa cập nhật thông tin ngân hàng
          </Text>
          <Text style={styles.noQRSubText}>
            Yêu cầu {toUserName} cập nhật tài khoản ngân hàng trong phần Hồ sơ
          </Text>
        </View>
      )}

      {/* Banking App Deeplinks */}
      {deeplinks.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            <Icon name="apps-outline" size={20} color={colors.text} /> Mở ứng dụng thanh toán
          </Text>
          <View style={styles.deeplinkGrid}>
            {deeplinks.map((deeplink, index) => (
              <TouchableOpacity
                key={index}
                style={[
                  styles.deeplinkCard,
                  {borderLeftColor: deeplink.color},
                ]}
                onPress={() => handleOpenDeeplink(deeplink)}>
                <View
                  style={[
                    styles.deeplinkIcon,
                    {backgroundColor: deeplink.color + '15'},
                  ]}>
                  <Icon
                    name={getDeeplinkIcon(deeplink.icon_name)}
                    size={24}
                    color={deeplink.color}
                  />
                </View>
                <Text style={styles.deeplinkName}>{deeplink.app_name}</Text>
                <Icon
                  name="open-outline"
                  size={16}
                  color={colors.textLight}
                />
              </TouchableOpacity>
            ))}
          </View>
        </View>
      )}

      {/* Bank Selection for QR */}
      {banks.length > 0 && paymentInfo?.bank_accounts?.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>
            <Icon name="business-outline" size={20} color={colors.text} /> Chọn ngân hàng tạo QR
          </Text>
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            contentContainerStyle={styles.bankScroll}>
            {banks.map(bank => (
              <TouchableOpacity
                key={bank.id}
                style={[
                  styles.bankChip,
                  selectedBank?.id === bank.id && styles.bankChipActive,
                  selectedBank?.id === bank.id && {
                    borderColor: bank.color,
                    backgroundColor: bank.color + '10',
                  },
                ]}
                onPress={() => handleGenerateQR(bank)}>
                <Image
                  source={{uri: bank.logo}}
                  style={styles.bankChipLogo}
                  resizeMode="contain"
                />
                <Text
                  style={[
                    styles.bankChipName,
                    selectedBank?.id === bank.id && {color: bank.color},
                  ]}>
                  {bank.short_name}
                </Text>
              </TouchableOpacity>
            ))}
          </ScrollView>
        </View>
      )}

      <View style={styles.bottomPadding} />
    </ScrollView>
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
  summaryCard: {
    backgroundColor: colors.surface,
    margin: spacing.md,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
    elevation: 3,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  summaryHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  summaryTitle: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginLeft: spacing.sm,
  },
  summaryAmount: {
    fontSize: 36,
    fontWeight: fontWeight.bold,
    color: colors.primary,
    textAlign: 'center',
    marginVertical: spacing.md,
  },
  summaryRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: spacing.xs,
  },
  summaryLabel: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
  },
  summaryValue: {
    fontSize: fontSize.md,
    fontWeight: fontWeight.medium,
    color: colors.text,
  },
  qrSection: {
    paddingHorizontal: spacing.md,
  },
  sectionTitle: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginBottom: spacing.md,
    marginTop: spacing.sm,
  },
  qrCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.lg,
    padding: spacing.lg,
    alignItems: 'center',
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.08,
    shadowRadius: 3,
  },
  bankHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  bankLogo: {
    width: 32,
    height: 32,
    marginRight: spacing.sm,
  },
  bankName: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
  },
  qrImage: {
    width: 280,
    height: 280,
    borderRadius: borderRadius.md,
  },
  accountInfo: {
    alignItems: 'center',
    marginTop: spacing.md,
    paddingTop: spacing.md,
    borderTopWidth: 1,
    borderTopColor: colors.borderLight,
    width: '100%',
  },
  accountNumber: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.bold,
    color: colors.text,
    letterSpacing: 1,
  },
  accountName: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    marginTop: spacing.xs,
  },
  qrHint: {
    fontSize: fontSize.sm,
    color: colors.textLight,
    textAlign: 'center',
    marginTop: spacing.md,
    fontStyle: 'italic',
  },
  noQRCard: {
    backgroundColor: colors.surface,
    margin: spacing.md,
    padding: spacing.xl,
    borderRadius: borderRadius.lg,
    alignItems: 'center',
    elevation: 2,
  },
  noQRText: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginTop: spacing.md,
    textAlign: 'center',
  },
  noQRSubText: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    marginTop: spacing.sm,
    textAlign: 'center',
  },
  section: {
    paddingHorizontal: spacing.md,
    marginTop: spacing.md,
  },
  deeplinkGrid: {
    gap: spacing.sm,
  },
  deeplinkCard: {
    backgroundColor: colors.surface,
    flexDirection: 'row',
    alignItems: 'center',
    padding: spacing.md,
    borderRadius: borderRadius.md,
    borderLeftWidth: 4,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
    marginBottom: spacing.sm,
  },
  deeplinkIcon: {
    width: 44,
    height: 44,
    borderRadius: borderRadius.md,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  deeplinkName: {
    flex: 1,
    fontSize: fontSize.lg,
    fontWeight: fontWeight.medium,
    color: colors.text,
  },
  bankScroll: {
    paddingBottom: spacing.sm,
  },
  bankChip: {
    backgroundColor: colors.surface,
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.full,
    borderWidth: 1.5,
    borderColor: colors.border,
    marginRight: spacing.sm,
    elevation: 1,
  },
  bankChipActive: {
    borderWidth: 2,
  },
  bankChipLogo: {
    width: 24,
    height: 24,
    marginRight: spacing.xs,
  },
  bankChipName: {
    fontSize: fontSize.sm,
    fontWeight: fontWeight.medium,
    color: colors.textSecondary,
  },
  bottomPadding: {
    height: spacing.xxl,
  },
});
