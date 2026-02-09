import React, {useState} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  ScrollView,
  Alert,
} from 'react-native';
import {useNavigation, useRoute, RouteProp} from '@react-navigation/native';
import Icon from 'react-native-vector-icons/Ionicons';
import {colors, spacing, fontSize, borderRadius} from '../../theme';
import {useBillStore} from '../../store/useBillStore';
import {useAuthStore} from '../../store/useAuthStore';
import {RootStackParamList} from '../../navigation/AppNavigator';
import {SplitType, CreateBillItemRequest} from '../../types';

type RouteProps = RouteProp<RootStackParamList, 'AddBill'>;

export default function AddBillScreen() {
  const navigation = useNavigation();
  const route = useRoute<RouteProps>();
  const {groupId, members} = route.params;
  const {createBill, isLoading} = useBillStore();
  const {user} = useAuthStore();

  const [title, setTitle] = useState('');
  const [category, setCategory] = useState('other');
  const [totalAmount, setTotalAmount] = useState('');
  const [splitType, setSplitType] = useState<SplitType>('equal');

  const categoryList = [
    {key: 'food', label: 'Ăn uống', icon: 'restaurant', color: '#FF6B6B'},
    {key: 'drinks', label: 'Đồ uống', icon: 'beer', color: '#FFA502'},
    {key: 'groceries', label: 'Tạp hóa', icon: 'cart', color: '#2ED573'},
    {key: 'transport', label: 'Di chuyển', icon: 'car', color: '#1E90FF'},
    {key: 'accommodation', label: 'Chỗ ở', icon: 'bed', color: '#A29BFE'},
    {key: 'entertainment', label: 'Giải trí', icon: 'game-controller', color: '#FD79A8'},
    {key: 'shopping', label: 'Mua sắm', icon: 'bag-handle', color: '#E17055'},
    {key: 'utilities', label: 'Tiện ích', icon: 'flash', color: '#FDCB6E'},
    {key: 'health', label: 'Sức khỏe', icon: 'medkit', color: '#00B894'},
    {key: 'travel', label: 'Du lịch', icon: 'airplane', color: '#74B9FF'},
    {key: 'other', label: 'Khác', icon: 'ellipsis-horizontal', color: '#636E72'},
  ];
  const [items, setItems] = useState<CreateBillItemRequest[]>([]);
  const [newItemName, setNewItemName] = useState('');
  const [newItemPrice, setNewItemPrice] = useState('');

  const addItem = () => {
    if (!newItemName || !newItemPrice) return;
    setItems([
      ...items,
      {
        name: newItemName,
        quantity: 1,
        unit_price: parseFloat(newItemPrice),
        total_price: parseFloat(newItemPrice),
        assigned_to: [],
      },
    ]);
    setNewItemName('');
    setNewItemPrice('');
  };

  const removeItem = (index: number) => {
    setItems(items.filter((_, i) => i !== index));
  };

  const handleCreate = async () => {
    if (!title.trim()) {
      Alert.alert('Lỗi', 'Vui lòng nhập tiêu đề');
      return;
    }
    if (!totalAmount || parseFloat(totalAmount) <= 0) {
      Alert.alert('Lỗi', 'Vui lòng nhập số tiền hợp lệ');
      return;
    }

    try {
      const memberIds = members.map((m: any) => m.user_id);
      await createBill(groupId, {
        title: title.trim(),
        category,
        total_amount: parseFloat(totalAmount),
        currency: 'VND',
        paid_by: user?.id || '',
        split_type: splitType,
        items: splitType === 'by_item' ? items : undefined,
        split_among: splitType === 'equal' ? memberIds : undefined,
        extra_charges: {tax: 0, service_charge: 0, tip: 0, discount: 0},
      });
      Alert.alert('Thành công', 'Hóa đơn đã được tạo!');
      navigation.goBack();
    } catch (error) {
      Alert.alert('Lỗi', 'Không thể tạo hóa đơn');
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView style={styles.form}>
        <Text style={styles.label}>Tiêu đề *</Text>
        <TextInput
          style={styles.input}
          placeholder="VD: Bữa tối nhà hàng ABC"
          value={title}
          onChangeText={setTitle}
        />

        <Text style={styles.label}>Danh mục</Text>
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          style={styles.categoryScroll}
          contentContainerStyle={styles.categoryContainer}>
          {categoryList.map(cat => (
            <TouchableOpacity
              key={cat.key}
              style={[
                styles.categoryChip,
                category === cat.key && {backgroundColor: cat.color},
              ]}
              onPress={() => setCategory(cat.key)}>
              <Icon
                name={cat.icon}
                size={16}
                color={category === cat.key ? '#FFF' : cat.color}
              />
              <Text
                style={[
                  styles.categoryChipText,
                  category === cat.key && {color: '#FFF'},
                ]}>
                {cat.label}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>

        <Text style={styles.label}>Tổng tiền (VNĐ) *</Text>
        <TextInput
          style={styles.input}
          placeholder="500000"
          keyboardType="numeric"
          value={totalAmount}
          onChangeText={setTotalAmount}
        />

        <Text style={styles.label}>Cách chia</Text>
        <View style={styles.splitOptions}>
          <TouchableOpacity
            style={[
              styles.splitOption,
              splitType === 'equal' && styles.splitOptionActive,
            ]}
            onPress={() => setSplitType('equal')}>
            <Icon
              name="people"
              size={20}
              color={splitType === 'equal' ? colors.textInverse : colors.primary}
            />
            <Text
              style={[
                styles.splitOptionText,
                splitType === 'equal' && styles.splitOptionTextActive,
              ]}>
              Chia đều
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[
              styles.splitOption,
              splitType === 'by_item' && styles.splitOptionActive,
            ]}
            onPress={() => setSplitType('by_item')}>
            <Icon
              name="list"
              size={20}
              color={splitType === 'by_item' ? colors.textInverse : colors.primary}
            />
            <Text
              style={[
                styles.splitOptionText,
                splitType === 'by_item' && styles.splitOptionTextActive,
              ]}>
              Theo món
            </Text>
          </TouchableOpacity>
        </View>

        {/* Equal Split Preview */}
        {splitType === 'equal' && totalAmount && (
          <View style={styles.previewCard}>
            <Text style={styles.previewTitle}>Mỗi người trả:</Text>
            <Text style={styles.previewAmount}>
              {(parseFloat(totalAmount) / (members.length || 1)).toLocaleString('vi-VN')}đ
            </Text>
            <Text style={styles.previewMeta}>
              ({members.length} người)
            </Text>
          </View>
        )}

        {/* By Item - Add Items */}
        {splitType === 'by_item' && (
          <View style={styles.itemsSection}>
            <Text style={styles.label}>Danh sách món</Text>
            {items.map((item, index) => (
              <View key={index} style={styles.itemRow}>
                <Text style={styles.itemName}>{item.name}</Text>
                <Text style={styles.itemPrice}>
                  {item.unit_price.toLocaleString('vi-VN')}đ
                </Text>
                <TouchableOpacity onPress={() => removeItem(index)}>
                  <Icon name="close-circle" size={20} color={colors.error} />
                </TouchableOpacity>
              </View>
            ))}

            <View style={styles.addItemRow}>
              <TextInput
                style={[styles.input, {flex: 2, marginRight: spacing.sm}]}
                placeholder="Tên món"
                value={newItemName}
                onChangeText={setNewItemName}
              />
              <TextInput
                style={[styles.input, {flex: 1, marginRight: spacing.sm}]}
                placeholder="Giá"
                keyboardType="numeric"
                value={newItemPrice}
                onChangeText={setNewItemPrice}
              />
              <TouchableOpacity style={styles.addItemBtn} onPress={addItem}>
                <Icon name="add" size={24} color={colors.textInverse} />
              </TouchableOpacity>
            </View>
          </View>
        )}

        <TouchableOpacity
          style={[styles.submitBtn, isLoading && styles.submitBtnDisabled]}
          onPress={handleCreate}
          disabled={isLoading}>
          <Text style={styles.submitBtnText}>
            {isLoading ? 'Đang tạo...' : 'Tạo Hóa Đơn'}
          </Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, backgroundColor: colors.background},
  form: {padding: spacing.lg},
  label: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
    marginBottom: spacing.sm,
    marginTop: spacing.md,
  },
  input: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.sm,
    borderWidth: 1,
    borderColor: colors.border,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    fontSize: fontSize.md,
    color: colors.text,
  },
  splitOptions: {
    flexDirection: 'row',
    gap: spacing.md,
  },
  splitOption: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: spacing.md,
    borderRadius: borderRadius.sm,
    borderWidth: 1,
    borderColor: colors.primary,
    gap: spacing.xs,
  },
  splitOptionActive: {
    backgroundColor: colors.primary,
  },
  splitOptionText: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.primary,
  },
  splitOptionTextActive: {
    color: colors.textInverse,
  },
  previewCard: {
    backgroundColor: colors.primaryLight + '15',
    borderRadius: borderRadius.md,
    padding: spacing.lg,
    alignItems: 'center',
    marginTop: spacing.md,
  },
  previewTitle: {fontSize: fontSize.md, color: colors.textSecondary},
  previewAmount: {
    fontSize: 32,
    fontWeight: '700',
    color: colors.primary,
    marginVertical: spacing.xs,
  },
  previewMeta: {fontSize: fontSize.sm, color: colors.textSecondary},
  itemsSection: {marginTop: spacing.md},
  itemRow: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: borderRadius.sm,
    padding: spacing.md,
    marginBottom: spacing.xs,
  },
  itemName: {flex: 1, fontSize: fontSize.md, color: colors.text},
  itemPrice: {fontSize: fontSize.md, fontWeight: '600', color: colors.primary, marginRight: spacing.sm},
  addItemRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: spacing.sm,
  },
  addItemBtn: {
    width: 44,
    height: 44,
    borderRadius: borderRadius.sm,
    backgroundColor: colors.secondary,
    justifyContent: 'center',
    alignItems: 'center',
  },
  categoryScroll: {
    marginBottom: spacing.xs,
  },
  categoryContainer: {
    gap: spacing.xs,
    paddingVertical: spacing.xs,
  },
  categoryChip: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: 20,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    gap: 4,
  },
  categoryChipText: {
    fontSize: fontSize.sm,
    fontWeight: '500',
    color: colors.text,
  },
  submitBtn: {
    backgroundColor: colors.primary,
    borderRadius: borderRadius.sm,
    paddingVertical: spacing.md,
    alignItems: 'center',
    marginTop: spacing.xl,
    marginBottom: spacing.xxl,
  },
  submitBtnDisabled: {opacity: 0.6},
  submitBtnText: {
    color: colors.textInverse,
    fontSize: fontSize.lg,
    fontWeight: '600',
  },
});
