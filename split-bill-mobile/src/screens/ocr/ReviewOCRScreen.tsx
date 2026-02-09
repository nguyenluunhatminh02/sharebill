import React, {useState} from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  TextInput,
  Alert,
  ActivityIndicator,
  ScrollView,
  FlatList,
} from 'react-native';
import {useNavigation, useRoute, RouteProp} from '@react-navigation/native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import Icon from 'react-native-vector-icons/Ionicons';

import {colors, spacing, borderRadius, fontSize, fontWeight} from '../../theme';
import {ocrAPI} from '../../api/services';
import {OCRResult, ParsedItem, ConfirmOCRRequest} from '../../types';
import {RootStackParamList} from '../../navigation/AppNavigator';
import {useAuthStore} from '../../store/useAuthStore';

type ReviewOCRRouteProp = RouteProp<RootStackParamList, 'ReviewOCR'>;
type NavigationProp = NativeStackNavigationProp<RootStackParamList>;

export default function ReviewOCRScreen() {
  const navigation = useNavigation<NavigationProp>();
  const route = useRoute<ReviewOCRRouteProp>();
  const {ocrResult, groupId, groupName} = route.params;
  const user = useAuthStore(state => state.user);

  const [title, setTitle] = useState(`Hóa đơn - ${groupName}`);
  const [items, setItems] = useState<ParsedItem[]>(ocrResult.parsed_items || []);
  const [total, setTotal] = useState(ocrResult.parsed_total);
  const [tax, setTax] = useState(ocrResult.parsed_tax);
  const [serviceFee, setServiceFee] = useState(ocrResult.parsed_service_fee);
  const [discount, setDiscount] = useState(ocrResult.parsed_discount);
  const [isConfirming, setIsConfirming] = useState(false);

  const formatVND = (amount: number) => {
    return amount.toLocaleString('vi-VN') + '₫';
  };

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.8) return colors.success;
    if (confidence >= 0.5) return colors.warning;
    return colors.error;
  };

  const handleUpdateItem = (index: number, field: keyof ParsedItem, value: string) => {
    const newItems = [...items];
    if (field === 'name') {
      newItems[index] = {...newItems[index], name: value};
    } else if (field === 'quantity') {
      const qty = parseInt(value) || 0;
      newItems[index] = {
        ...newItems[index],
        quantity: qty,
        total_price: qty * newItems[index].unit_price,
      };
    } else if (field === 'unit_price') {
      const price = parseFloat(value) || 0;
      newItems[index] = {
        ...newItems[index],
        unit_price: price,
        total_price: newItems[index].quantity * price,
      };
    } else if (field === 'total_price') {
      newItems[index] = {
        ...newItems[index],
        total_price: parseFloat(value) || 0,
      };
    }
    setItems(newItems);
  };

  const handleRemoveItem = (index: number) => {
    const newItems = items.filter((_, i) => i !== index);
    setItems(newItems);
  };

  const handleAddItem = () => {
    setItems([
      ...items,
      {
        name: 'Món mới',
        quantity: 1,
        unit_price: 0,
        total_price: 0,
        confidence: 1.0,
      },
    ]);
  };

  const calculateItemsTotal = () => {
    return items.reduce((sum, item) => sum + item.total_price, 0);
  };

  const handleConfirm = async () => {
    if (!title.trim()) {
      Alert.alert('Lỗi', 'Vui lòng nhập tên hóa đơn');
      return;
    }
    if (items.length === 0) {
      Alert.alert('Lỗi', 'Hóa đơn phải có ít nhất 1 món');
      return;
    }
    if (total <= 0) {
      Alert.alert('Lỗi', 'Tổng tiền phải lớn hơn 0');
      return;
    }

    setIsConfirming(true);
    try {
      const confirmData: ConfirmOCRRequest = {
        title,
        items,
        total,
        tax,
        service_fee: serviceFee,
        discount,
        paid_by: user?.id || '',
        split_type: 'equal',
        split_among: [],
      };

      const response = await ocrAPI.confirm(ocrResult.id, confirmData);

      Alert.alert(
        'Thành công! ✅',
        'Hóa đơn đã được tạo từ ảnh chụp',
        [
          {
            text: 'Xem hóa đơn',
            onPress: () => {
              navigation.navigate('BillDetail', {
                billId: response.data.data?.id || '',
              });
            },
          },
          {
            text: 'Về nhóm',
            onPress: () => {
              navigation.navigate('GroupDetail', {
                groupId,
                groupName,
              });
            },
          },
        ],
      );
    } catch (error: any) {
      const errorMsg =
        error?.response?.data?.error || 'Đã xảy ra lỗi khi xác nhận';
      Alert.alert('Lỗi', errorMsg);
    } finally {
      setIsConfirming(false);
    }
  };

  const renderItem = ({item, index}: {item: ParsedItem; index: number}) => (
    <View style={styles.itemCard}>
      <View style={styles.itemHeader}>
        <View style={styles.itemConfidence}>
          <View
            style={[
              styles.confidenceDot,
              {backgroundColor: getConfidenceColor(item.confidence)},
            ]}
          />
          <Text style={styles.confidenceText}>
            {Math.round(item.confidence * 100)}%
          </Text>
        </View>
        <TouchableOpacity onPress={() => handleRemoveItem(index)}>
          <Icon name="trash-outline" size={20} color={colors.error} />
        </TouchableOpacity>
      </View>

      <TextInput
        style={styles.itemNameInput}
        value={item.name}
        onChangeText={value => handleUpdateItem(index, 'name', value)}
        placeholder="Tên món"
      />

      <View style={styles.itemDetailsRow}>
        <View style={styles.itemField}>
          <Text style={styles.itemFieldLabel}>SL</Text>
          <TextInput
            style={styles.itemFieldInput}
            value={item.quantity.toString()}
            onChangeText={value => handleUpdateItem(index, 'quantity', value)}
            keyboardType="numeric"
          />
        </View>
        <View style={styles.itemField}>
          <Text style={styles.itemFieldLabel}>Đơn giá</Text>
          <TextInput
            style={styles.itemFieldInput}
            value={item.unit_price.toString()}
            onChangeText={value => handleUpdateItem(index, 'unit_price', value)}
            keyboardType="numeric"
          />
        </View>
        <View style={styles.itemField}>
          <Text style={styles.itemFieldLabel}>Thành tiền</Text>
          <Text style={styles.itemTotalText}>{formatVND(item.total_price)}</Text>
        </View>
      </View>
    </View>
  );

  return (
    <View style={styles.container}>
      <ScrollView contentContainerStyle={styles.content}>
        {/* Confidence Score */}
        <View style={styles.confidenceCard}>
          <View style={styles.confidenceRow}>
            <Icon name="analytics" size={22} color={colors.primary} />
            <Text style={styles.confidenceLabel}>Độ chính xác:</Text>
            <Text
              style={[
                styles.confidenceValue,
                {color: getConfidenceColor(ocrResult.confidence_score)},
              ]}>
              {Math.round(ocrResult.confidence_score * 100)}%
            </Text>
          </View>
          <Text style={styles.processingTime}>
            Thời gian xử lý: {ocrResult.processing_time_ms}ms
          </Text>
        </View>

        {/* Bill Title */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Tên hóa đơn</Text>
          <TextInput
            style={styles.titleInput}
            value={title}
            onChangeText={setTitle}
            placeholder="Nhập tên hóa đơn"
          />
        </View>

        {/* Items */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>
              Các món ({items.length})
            </Text>
            <TouchableOpacity style={styles.addItemBtn} onPress={handleAddItem}>
              <Icon name="add-circle" size={22} color={colors.primary} />
              <Text style={styles.addItemText}>Thêm món</Text>
            </TouchableOpacity>
          </View>

          {items.map((item, index) => (
            <View key={index}>{renderItem({item, index})}</View>
          ))}

          {items.length === 0 && (
            <View style={styles.emptyItems}>
              <Icon name="alert-circle-outline" size={32} color={colors.warning} />
              <Text style={styles.emptyItemsText}>
                Không tìm thấy món nào. Vui lòng thêm thủ công.
              </Text>
            </View>
          )}
        </View>

        {/* Totals */}
        <View style={styles.totalsCard}>
          <Text style={styles.totalsTitle}>Tổng kết</Text>

          <View style={styles.totalRow}>
            <Text style={styles.totalLabel}>Tổng các món:</Text>
            <Text style={styles.totalValue}>{formatVND(calculateItemsTotal())}</Text>
          </View>

          <View style={styles.totalInputRow}>
            <Text style={styles.totalLabel}>Thuế (VAT):</Text>
            <TextInput
              style={styles.totalInput}
              value={tax.toString()}
              onChangeText={v => setTax(parseFloat(v) || 0)}
              keyboardType="numeric"
            />
          </View>

          <View style={styles.totalInputRow}>
            <Text style={styles.totalLabel}>Phí phục vụ:</Text>
            <TextInput
              style={styles.totalInput}
              value={serviceFee.toString()}
              onChangeText={v => setServiceFee(parseFloat(v) || 0)}
              keyboardType="numeric"
            />
          </View>

          <View style={styles.totalInputRow}>
            <Text style={styles.totalLabel}>Giảm giá:</Text>
            <TextInput
              style={styles.totalInput}
              value={discount.toString()}
              onChangeText={v => setDiscount(parseFloat(v) || 0)}
              keyboardType="numeric"
            />
          </View>

          <View style={styles.divider} />

          <View style={styles.totalInputRow}>
            <Text style={styles.grandTotalLabel}>Tổng cộng:</Text>
            <TextInput
              style={[styles.totalInput, styles.grandTotalInput]}
              value={total.toString()}
              onChangeText={v => setTotal(parseFloat(v) || 0)}
              keyboardType="numeric"
            />
          </View>

          <TouchableOpacity
            style={styles.autoCalcButton}
            onPress={() => {
              const autoTotal = calculateItemsTotal() + tax + serviceFee - discount;
              setTotal(autoTotal);
            }}>
            <Icon name="calculator-outline" size={16} color={colors.primary} />
            <Text style={styles.autoCalcText}>Tự động tính tổng</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>

      {/* Bottom confirm button */}
      <View style={styles.bottomBar}>
        <TouchableOpacity
          style={[styles.confirmButton, isConfirming && styles.confirmButtonDisabled]}
          onPress={handleConfirm}
          disabled={isConfirming}>
          {isConfirming ? (
            <ActivityIndicator size="small" color={colors.textInverse} />
          ) : (
            <>
              <Icon name="checkmark-circle" size={22} color={colors.textInverse} />
              <Text style={styles.confirmButtonText}>
                Xác nhận & Tạo hóa đơn ({formatVND(total)})
              </Text>
            </>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    padding: spacing.md,
    paddingBottom: 100,
  },
  confidenceCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.md,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  confidenceRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  confidenceLabel: {
    fontSize: fontSize.lg,
    color: colors.text,
    fontWeight: fontWeight.medium,
  },
  confidenceValue: {
    fontSize: fontSize.xl,
    fontWeight: fontWeight.bold,
  },
  processingTime: {
    fontSize: fontSize.sm,
    color: colors.textLight,
    marginTop: spacing.xs,
  },
  section: {
    marginBottom: spacing.md,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  sectionTitle: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginBottom: spacing.sm,
  },
  titleInput: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.sm,
    padding: spacing.md,
    fontSize: fontSize.lg,
    color: colors.text,
    borderWidth: 1,
    borderColor: colors.border,
  },
  addItemBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  addItemText: {
    fontSize: fontSize.md,
    color: colors.primary,
    fontWeight: fontWeight.medium,
  },
  itemCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.sm,
    padding: spacing.md,
    marginBottom: spacing.sm,
    borderWidth: 1,
    borderColor: colors.borderLight,
  },
  itemHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  itemConfidence: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  confidenceDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  confidenceText: {
    fontSize: fontSize.xs,
    color: colors.textLight,
  },
  itemNameInput: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.medium,
    color: colors.text,
    borderBottomWidth: 1,
    borderBottomColor: colors.borderLight,
    paddingBottom: spacing.xs,
    marginBottom: spacing.sm,
  },
  itemDetailsRow: {
    flexDirection: 'row',
    gap: spacing.sm,
  },
  itemField: {
    flex: 1,
  },
  itemFieldLabel: {
    fontSize: fontSize.xs,
    color: colors.textLight,
    marginBottom: 2,
  },
  itemFieldInput: {
    fontSize: fontSize.md,
    color: colors.text,
    backgroundColor: colors.background,
    borderRadius: borderRadius.sm,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    textAlign: 'center',
  },
  itemTotalText: {
    fontSize: fontSize.md,
    color: colors.primary,
    fontWeight: fontWeight.semibold,
    textAlign: 'center',
    paddingVertical: spacing.xs,
  },
  emptyItems: {
    alignItems: 'center',
    padding: spacing.lg,
    gap: spacing.sm,
  },
  emptyItemsText: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    textAlign: 'center',
  },
  totalsCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  totalsTitle: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginBottom: spacing.md,
  },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  totalLabel: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
  },
  totalValue: {
    fontSize: fontSize.md,
    color: colors.text,
    fontWeight: fontWeight.medium,
  },
  totalInputRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  totalInput: {
    backgroundColor: colors.background,
    borderRadius: borderRadius.sm,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    fontSize: fontSize.md,
    color: colors.text,
    textAlign: 'right',
    minWidth: 120,
    borderWidth: 1,
    borderColor: colors.borderLight,
  },
  grandTotalLabel: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.bold,
    color: colors.text,
  },
  grandTotalInput: {
    fontWeight: fontWeight.bold,
    fontSize: fontSize.lg,
    borderColor: colors.primary,
  },
  divider: {
    height: 1,
    backgroundColor: colors.border,
    marginVertical: spacing.sm,
  },
  autoCalcButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 4,
    paddingVertical: spacing.sm,
    marginTop: spacing.xs,
  },
  autoCalcText: {
    fontSize: fontSize.sm,
    color: colors.primary,
    fontWeight: fontWeight.medium,
  },
  bottomBar: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: colors.surface,
    padding: spacing.md,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    elevation: 10,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: -2},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  confirmButton: {
    backgroundColor: colors.success,
    borderRadius: borderRadius.md,
    paddingVertical: spacing.md,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
  },
  confirmButtonDisabled: {
    backgroundColor: colors.textLight,
  },
  confirmButtonText: {
    color: colors.textInverse,
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
  },
});
