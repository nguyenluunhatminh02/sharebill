import React, {useEffect, useState} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  RefreshControl,
  TextInput,
  Alert,
} from 'react-native';
import {useNavigation} from '@react-navigation/native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import Icon from 'react-native-vector-icons/Ionicons';
import {colors, spacing, fontSize, borderRadius} from '../../theme';
import {useGroupStore} from '../../store/useGroupStore';
import {RootStackParamList} from '../../navigation/AppNavigator';
import {Group} from '../../types';

type NavigationProp = NativeStackNavigationProp<RootStackParamList>;

export default function GroupListScreen() {
  const navigation = useNavigation<NavigationProp>();
  const {groups, fetchGroups, joinGroup, isLoading} = useGroupStore();
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [inviteCode, setInviteCode] = useState('');

  useEffect(() => {
    fetchGroups();
  }, []);

  const handleJoinGroup = async () => {
    if (!inviteCode.trim()) return;
    try {
      await joinGroup(inviteCode.trim());
      setShowJoinModal(false);
      setInviteCode('');
      Alert.alert('Thành công', 'Đã tham gia nhóm!');
    } catch (error) {
      Alert.alert('Lỗi', 'Mã mời không hợp lệ');
    }
  };

  const renderGroup = ({item}: {item: Group}) => (
    <TouchableOpacity
      style={styles.groupCard}
      onPress={() =>
        navigation.navigate('GroupDetail', {
          groupId: item.id,
          groupName: item.name,
        })
      }>
      <View style={styles.avatar}>
        <Text style={styles.avatarText}>
          {item.name.charAt(0).toUpperCase()}
        </Text>
      </View>
      <View style={styles.groupInfo}>
        <Text style={styles.groupName}>{item.name}</Text>
        <Text style={styles.groupDesc}>
          {item.members?.length || 0} thành viên
          {item.description ? ` • ${item.description}` : ''}
        </Text>
      </View>
      <Icon name="chevron-forward" size={20} color={colors.textLight} />
    </TouchableOpacity>
  );

  return (
    <SafeAreaView style={styles.container}>
      {/* Header Actions */}
      <View style={styles.header}>
        <Text style={styles.title}>Nhóm Của Tôi</Text>
        <View style={styles.headerActions}>
          <TouchableOpacity
            style={styles.headerBtn}
            onPress={() => setShowJoinModal(!showJoinModal)}>
            <Icon name="enter-outline" size={22} color={colors.primary} />
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.headerBtn}
            onPress={() => navigation.navigate('CreateGroup')}>
            <Icon name="add" size={24} color={colors.primary} />
          </TouchableOpacity>
        </View>
      </View>

      {/* Join Group Input */}
      {showJoinModal && (
        <View style={styles.joinContainer}>
          <TextInput
            style={styles.joinInput}
            placeholder="Nhập mã mời nhóm..."
            value={inviteCode}
            onChangeText={setInviteCode}
            autoCapitalize="characters"
          />
          <TouchableOpacity style={styles.joinButton} onPress={handleJoinGroup}>
            <Text style={styles.joinButtonText}>Tham gia</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Group List */}
      <FlatList
        data={groups}
        renderItem={renderGroup}
        keyExtractor={item => item.id}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl refreshing={isLoading} onRefresh={fetchGroups} />
        }
        ListEmptyComponent={
          <View style={styles.emptyState}>
            <Icon name="people-outline" size={64} color={colors.textLight} />
            <Text style={styles.emptyTitle}>Chưa có nhóm nào</Text>
            <Text style={styles.emptyDesc}>
              Tạo nhóm mới hoặc tham gia bằng mã mời
            </Text>
          </View>
        }
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, backgroundColor: colors.background},
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
  },
  title: {fontSize: fontSize.xxl, fontWeight: '700', color: colors.text},
  headerActions: {flexDirection: 'row', gap: spacing.sm},
  headerBtn: {
    width: 40,
    height: 40,
    borderRadius: borderRadius.full,
    backgroundColor: colors.primaryLight + '20',
    justifyContent: 'center',
    alignItems: 'center',
  },
  joinContainer: {
    flexDirection: 'row',
    paddingHorizontal: spacing.lg,
    marginBottom: spacing.md,
    gap: spacing.sm,
  },
  joinInput: {
    flex: 1,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.sm,
    paddingHorizontal: spacing.md,
    fontSize: fontSize.md,
    borderWidth: 1,
    borderColor: colors.border,
  },
  joinButton: {
    backgroundColor: colors.primary,
    borderRadius: borderRadius.sm,
    paddingHorizontal: spacing.lg,
    justifyContent: 'center',
  },
  joinButtonText: {color: colors.textInverse, fontWeight: '600'},
  listContent: {paddingHorizontal: spacing.lg, paddingBottom: spacing.xxl},
  groupCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 4,
  },
  avatar: {
    width: 52,
    height: 52,
    borderRadius: borderRadius.full,
    backgroundColor: colors.primaryLight + '30',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  avatarText: {fontSize: fontSize.xl, fontWeight: '700', color: colors.primary},
  groupInfo: {flex: 1},
  groupName: {fontSize: fontSize.lg, fontWeight: '600', color: colors.text},
  groupDesc: {fontSize: fontSize.sm, color: colors.textSecondary, marginTop: 2},
  emptyState: {
    alignItems: 'center',
    paddingTop: 100,
  },
  emptyTitle: {
    fontSize: fontSize.xl,
    fontWeight: '600',
    color: colors.text,
    marginTop: spacing.md,
  },
  emptyDesc: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    marginTop: spacing.xs,
    textAlign: 'center',
  },
});
