import React, {useState} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  KeyboardAvoidingView,
  Platform,
  Alert,
} from 'react-native';
import Icon from 'react-native-vector-icons/Ionicons';
import {colors, spacing, fontSize, borderRadius} from '../../theme';
import {useAuthStore} from '../../store/useAuthStore';

export default function LoginScreen() {
  const [phone, setPhone] = useState('');
  const [otp, setOtp] = useState('');
  const [step, setStep] = useState<'phone' | 'otp'>('phone');
  const [isLoading, setIsLoading] = useState(false);
  const {setToken, verifyToken} = useAuthStore();

  const handleSendOTP = async () => {
    if (phone.length < 9) {
      Alert.alert('Lỗi', 'Vui lòng nhập số điện thoại hợp lệ');
      return;
    }
    setIsLoading(true);

    try {
      // TODO: Integrate Firebase Phone Auth
      // const confirmation = await auth().signInWithPhoneNumber('+84' + phone);
      // setConfirmation(confirmation);

      // For development: simulate OTP sent
      setTimeout(() => {
        setStep('otp');
        setIsLoading(false);
        Alert.alert('Dev Mode', 'OTP đã gửi (dev: dùng bất kỳ 6 số)');
      }, 1000);
    } catch (error) {
      Alert.alert('Lỗi', 'Không thể gửi OTP. Vui lòng thử lại.');
      setIsLoading(false);
    }
  };

  const handleVerifyOTP = async () => {
    if (otp.length !== 6) {
      Alert.alert('Lỗi', 'Vui lòng nhập mã OTP 6 số');
      return;
    }
    setIsLoading(true);

    try {
      // TODO: Integrate Firebase Phone Auth verification
      // const credential = await confirmation.confirm(otp);
      // const token = await credential.user.getIdToken();

      // For development: use phone as token
      const devToken = 'dev-user-' + phone;
      setToken(devToken);
      await verifyToken();
    } catch (error) {
      Alert.alert('Lỗi', 'Mã OTP không đúng. Vui lòng thử lại.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.content}>
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.iconContainer}>
            <Icon name="receipt-outline" size={48} color={colors.primary} />
          </View>
          <Text style={styles.title}>Split Bill</Text>
          <Text style={styles.subtitle}>Chia tiền nhóm thông minh</Text>
        </View>

        {/* Form */}
        <View style={styles.form}>
          {step === 'phone' ? (
            <>
              <Text style={styles.label}>Số điện thoại</Text>
              <View style={styles.phoneInput}>
                <Text style={styles.countryCode}>+84</Text>
                <TextInput
                  style={styles.input}
                  placeholder="912 345 678"
                  keyboardType="phone-pad"
                  value={phone}
                  onChangeText={setPhone}
                  maxLength={10}
                />
              </View>
              <TouchableOpacity
                style={[styles.button, isLoading && styles.buttonDisabled]}
                onPress={handleSendOTP}
                disabled={isLoading}>
                <Text style={styles.buttonText}>
                  {isLoading ? 'Đang gửi...' : 'Gửi mã OTP'}
                </Text>
              </TouchableOpacity>
            </>
          ) : (
            <>
              <Text style={styles.label}>Nhập mã OTP</Text>
              <Text style={styles.otpInfo}>
                Mã đã gửi đến +84{phone}
              </Text>
              <TextInput
                style={styles.otpInput}
                placeholder="000000"
                keyboardType="number-pad"
                value={otp}
                onChangeText={setOtp}
                maxLength={6}
                textAlign="center"
              />
              <TouchableOpacity
                style={[styles.button, isLoading && styles.buttonDisabled]}
                onPress={handleVerifyOTP}
                disabled={isLoading}>
                <Text style={styles.buttonText}>
                  {isLoading ? 'Đang xác thực...' : 'Xác nhận'}
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                onPress={() => setStep('phone')}
                style={styles.backButton}>
                <Text style={styles.backButtonText}>← Đổi số điện thoại</Text>
              </TouchableOpacity>
            </>
          )}
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: spacing.xl,
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing.xxl,
  },
  iconContainer: {
    width: 96,
    height: 96,
    borderRadius: borderRadius.xl,
    backgroundColor: colors.primaryLight + '20',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  title: {
    fontSize: fontSize.xxxl,
    fontWeight: '700',
    color: colors.primary,
    marginBottom: spacing.xs,
  },
  subtitle: {
    fontSize: fontSize.lg,
    color: colors.textSecondary,
  },
  form: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.lg,
    padding: spacing.lg,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 4,
  },
  label: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
    marginBottom: spacing.sm,
  },
  phoneInput: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: borderRadius.sm,
    marginBottom: spacing.md,
  },
  countryCode: {
    paddingHorizontal: spacing.md,
    fontSize: fontSize.lg,
    fontWeight: '600',
    color: colors.text,
    borderRightWidth: 1,
    borderRightColor: colors.border,
    paddingVertical: spacing.md,
  },
  input: {
    flex: 1,
    paddingHorizontal: spacing.md,
    fontSize: fontSize.lg,
    color: colors.text,
    paddingVertical: spacing.md,
  },
  otpInfo: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginBottom: spacing.md,
  },
  otpInput: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: borderRadius.sm,
    paddingVertical: spacing.md,
    fontSize: 28,
    fontWeight: '700',
    letterSpacing: 8,
    color: colors.text,
    marginBottom: spacing.md,
  },
  button: {
    backgroundColor: colors.primary,
    borderRadius: borderRadius.sm,
    paddingVertical: spacing.md,
    alignItems: 'center',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    color: colors.textInverse,
    fontSize: fontSize.lg,
    fontWeight: '600',
  },
  backButton: {
    marginTop: spacing.md,
    alignItems: 'center',
  },
  backButtonText: {
    color: colors.primary,
    fontSize: fontSize.md,
  },
});
