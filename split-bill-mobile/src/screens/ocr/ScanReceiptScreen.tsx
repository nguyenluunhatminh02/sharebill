import React, {useState} from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Image,
  Alert,
  ActivityIndicator,
  ScrollView,
  Platform,
} from 'react-native';
import {useNavigation, useRoute, RouteProp} from '@react-navigation/native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import Icon from 'react-native-vector-icons/Ionicons';
import {launchCamera, launchImageLibrary, Asset} from 'react-native-image-picker';

import {colors, spacing, borderRadius, fontSize, fontWeight} from '../../theme';
import {ocrAPI, uploadAPI} from '../../api/services';
import {OCRResult} from '../../types';
import {RootStackParamList} from '../../navigation/AppNavigator';

type ScanReceiptRouteProp = RouteProp<RootStackParamList, 'ScanReceipt'>;
type NavigationProp = NativeStackNavigationProp<RootStackParamList>;

export default function ScanReceiptScreen() {
  const navigation = useNavigation<NavigationProp>();
  const route = useRoute<ScanReceiptRouteProp>();
  const {groupId, groupName} = route.params;

  const [selectedImage, setSelectedImage] = useState<Asset | null>(null);
  const [isScanning, setIsScanning] = useState(false);
  const [scanProgress, setScanProgress] = useState('');

  const handlePickImage = async (source: 'camera' | 'gallery') => {
    const options = {
      mediaType: 'photo' as const,
      quality: 0.8 as const,
      maxWidth: 2048,
      maxHeight: 2048,
      includeBase64: true,
    };

    try {
      const result =
        source === 'camera'
          ? await launchCamera(options)
          : await launchImageLibrary(options);

      if (result.didCancel) return;
      if (result.errorCode) {
        Alert.alert('Lỗi', result.errorMessage || 'Không thể truy cập ảnh');
        return;
      }

      if (result.assets && result.assets[0]) {
        setSelectedImage(result.assets[0]);
      }
    } catch (error) {
      Alert.alert('Lỗi', 'Không thể chọn ảnh');
    }
  };

  const handleScanReceipt = async () => {
    if (!selectedImage?.base64) {
      Alert.alert('Lỗi', 'Vui lòng chọn ảnh hóa đơn');
      return;
    }

    setIsScanning(true);
    setScanProgress('Đang tải ảnh lên...');

    try {
      // Step 1: Upload image
      const imageData = `data:${selectedImage.type || 'image/jpeg'};base64,${selectedImage.base64}`;

      setScanProgress('Đang phân tích hóa đơn...');

      // Step 2: Scan receipt using base64
      const scanResponse = await ocrAPI.scanReceiptBase64({
        group_id: groupId,
        image_base64: imageData,
        file_name: selectedImage.fileName,
      });

      const ocrResult = scanResponse.data.data;

      if (!ocrResult) {
        Alert.alert('Lỗi', 'Không thể phân tích hóa đơn');
        return;
      }

      setScanProgress('Hoàn thành!');

      // Navigate to review screen
      navigation.navigate('ReviewOCR', {
        ocrResult,
        groupId,
        groupName,
      });
    } catch (error: any) {
      const errorMsg =
        error?.response?.data?.error || 'Đã xảy ra lỗi khi quét hóa đơn';
      Alert.alert('Lỗi quét hóa đơn', errorMsg);
    } finally {
      setIsScanning(false);
      setScanProgress('');
    }
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* Header instructions */}
      <View style={styles.instructionCard}>
        <Icon name="information-circle" size={24} color={colors.primary} />
        <Text style={styles.instructionText}>
          Chụp hoặc chọn ảnh hóa đơn để tự động nhận diện các món và tổng tiền.
          Hệ thống hỗ trợ hóa đơn tiếng Việt.
        </Text>
      </View>

      {/* Image selection area */}
      <View style={styles.imageSection}>
        {selectedImage ? (
          <View style={styles.previewContainer}>
            <Image
              source={{uri: selectedImage.uri}}
              style={styles.previewImage}
              resizeMode="contain"
            />
            <TouchableOpacity
              style={styles.removeImageBtn}
              onPress={() => setSelectedImage(null)}>
              <Icon name="close-circle" size={28} color={colors.error} />
            </TouchableOpacity>
          </View>
        ) : (
          <View style={styles.placeholderContainer}>
            <Icon name="receipt-outline" size={64} color={colors.textLight} />
            <Text style={styles.placeholderText}>Chọn ảnh hóa đơn</Text>
          </View>
        )}
      </View>

      {/* Action buttons */}
      <View style={styles.actionRow}>
        <TouchableOpacity
          style={styles.actionButton}
          onPress={() => handlePickImage('camera')}
          disabled={isScanning}>
          <View style={[styles.actionIconWrapper, {backgroundColor: colors.primaryLight + '20'}]}>
            <Icon name="camera" size={28} color={colors.primary} />
          </View>
          <Text style={styles.actionLabel}>Chụp ảnh</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.actionButton}
          onPress={() => handlePickImage('gallery')}
          disabled={isScanning}>
          <View style={[styles.actionIconWrapper, {backgroundColor: colors.secondaryLight + '20'}]}>
            <Icon name="images" size={28} color={colors.secondary} />
          </View>
          <Text style={styles.actionLabel}>Thư viện</Text>
        </TouchableOpacity>
      </View>

      {/* Scan button */}
      <TouchableOpacity
        style={[
          styles.scanButton,
          (!selectedImage || isScanning) && styles.scanButtonDisabled,
        ]}
        onPress={handleScanReceipt}
        disabled={!selectedImage || isScanning}>
        {isScanning ? (
          <View style={styles.scanningRow}>
            <ActivityIndicator size="small" color={colors.textInverse} />
            <Text style={styles.scanButtonText}>{scanProgress}</Text>
          </View>
        ) : (
          <View style={styles.scanningRow}>
            <Icon name="scan" size={22} color={colors.textInverse} />
            <Text style={styles.scanButtonText}>Quét Hóa Đơn</Text>
          </View>
        )}
      </TouchableOpacity>

      {/* Tips */}
      <View style={styles.tipsCard}>
        <Text style={styles.tipsTitle}>💡 Mẹo chụp ảnh</Text>
        <View style={styles.tipRow}>
          <Icon name="checkmark-circle" size={16} color={colors.success} />
          <Text style={styles.tipText}>Đặt hóa đơn trên nền phẳng, sáng màu</Text>
        </View>
        <View style={styles.tipRow}>
          <Icon name="checkmark-circle" size={16} color={colors.success} />
          <Text style={styles.tipText}>Chụp thẳng từ trên xuống, không nghiêng</Text>
        </View>
        <View style={styles.tipRow}>
          <Icon name="checkmark-circle" size={16} color={colors.success} />
          <Text style={styles.tipText}>Đảm bảo đủ ánh sáng, không bị mờ</Text>
        </View>
        <View style={styles.tipRow}>
          <Icon name="checkmark-circle" size={16} color={colors.success} />
          <Text style={styles.tipText}>Bao gồm toàn bộ hóa đơn trong khung hình</Text>
        </View>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    padding: spacing.md,
    paddingBottom: spacing.xxl,
  },
  instructionCard: {
    flexDirection: 'row',
    backgroundColor: colors.primaryLight + '15',
    padding: spacing.md,
    borderRadius: borderRadius.md,
    marginBottom: spacing.md,
    alignItems: 'flex-start',
    gap: spacing.sm,
  },
  instructionText: {
    flex: 1,
    fontSize: fontSize.md,
    color: colors.text,
    lineHeight: 20,
  },
  imageSection: {
    marginBottom: spacing.md,
  },
  previewContainer: {
    borderRadius: borderRadius.md,
    overflow: 'hidden',
    backgroundColor: colors.surface,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  previewImage: {
    width: '100%',
    height: 350,
    backgroundColor: '#f0f0f0',
  },
  removeImageBtn: {
    position: 'absolute',
    top: spacing.sm,
    right: spacing.sm,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.full,
    padding: 2,
  },
  placeholderContainer: {
    height: 250,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    borderWidth: 2,
    borderColor: colors.border,
    borderStyle: 'dashed',
    justifyContent: 'center',
    alignItems: 'center',
  },
  placeholderText: {
    marginTop: spacing.sm,
    fontSize: fontSize.lg,
    color: colors.textLight,
  },
  actionRow: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: spacing.xl,
    marginBottom: spacing.lg,
  },
  actionButton: {
    alignItems: 'center',
  },
  actionIconWrapper: {
    width: 64,
    height: 64,
    borderRadius: borderRadius.lg,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.xs,
  },
  actionLabel: {
    fontSize: fontSize.md,
    color: colors.text,
    fontWeight: fontWeight.medium,
  },
  scanButton: {
    backgroundColor: colors.primary,
    borderRadius: borderRadius.md,
    paddingVertical: spacing.md,
    alignItems: 'center',
    marginBottom: spacing.lg,
    elevation: 3,
    shadowColor: colors.primary,
    shadowOffset: {width: 0, height: 4},
    shadowOpacity: 0.3,
    shadowRadius: 8,
  },
  scanButtonDisabled: {
    backgroundColor: colors.textLight,
    elevation: 0,
    shadowOpacity: 0,
  },
  scanButtonText: {
    color: colors.textInverse,
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
  },
  scanningRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  tipsCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  tipsTitle: {
    fontSize: fontSize.lg,
    fontWeight: fontWeight.semibold,
    color: colors.text,
    marginBottom: spacing.sm,
  },
  tipRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    marginBottom: spacing.xs,
  },
  tipText: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
  },
});
